package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// DLQMessage wraps the original message with failure metadata so
// operators can inspect, diagnose, and optionally replay it.
type DLQMessage struct {
	OriginalTopic     string          `json:"original_topic"`
	OriginalPartition int             `json:"original_partition"`
	OriginalOffset    int64           `json:"original_offset"`
	OriginalKey       string          `json:"original_key"`
	OriginalValue     json.RawMessage `json:"original_value"`
	ErrorMessage      string          `json:"error_message"`
	ErrorKind         string          `json:"error_kind"` // "transient" | "permanent"
	Retries           int             `json:"retries"`
	FailedAt          time.Time       `json:"failed_at"`
}

// DLQProducer writes failed messages to a dead-letter topic.
type DLQProducer struct {
	writer *kafka.Writer
	topic  string
}

// NewDLQProducer creates a producer targeting the dead-letter topic.
func NewDLQProducer(brokers []string, dlqTopic string) *DLQProducer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        dlqTopic,
		RequiredAcks: kafka.RequireAll,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
	}
	return &DLQProducer{writer: w, topic: dlqTopic}
}

// Send routes a failed message to the DLQ with full failure context.
func (d *DLQProducer) Send(ctx context.Context, original kafka.Message, reason error, kind ErrorKind, retries int) error {
	envelope := DLQMessage{
		OriginalTopic:     original.Topic,
		OriginalPartition: original.Partition,
		OriginalOffset:    original.Offset,
		OriginalKey:       string(original.Key),
		OriginalValue:     original.Value,
		ErrorMessage:      reason.Error(),
		ErrorKind:         kind.String(),
		Retries:           retries,
		FailedAt:          time.Now(),
	}

	value, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal DLQ envelope: %w", err)
	}

	msg := kafka.Message{
		Key:   original.Key,
		Value: value,
		Headers: []kafka.Header{
			{Key: "dlq-reason", Value: []byte(kind.String())},
			{Key: "original-topic", Value: []byte(original.Topic)},
		},
	}

	if err := d.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("write to DLQ topic %s: %w", d.topic, err)
	}
	return nil
}

// Close shuts down the DLQ writer.
func (d *DLQProducer) Close() error {
	return d.writer.Close()
}

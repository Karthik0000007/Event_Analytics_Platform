package messaging

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// Consumer wraps kafka-go Reader with manual commit control.
// Auto-commit is disabled so offsets are only committed after a
// successful DB write, guaranteeing at-least-once delivery.
type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       10e6, // 10 MB
		MaxWait:        1 * time.Second,
		CommitInterval: 0, // disable auto-commit
		StartOffset:    kafka.LastOffset,
	})

	return &Consumer{reader: r}
}

// FetchMessage retrieves the next message without committing the offset.
func (c *Consumer) FetchMessage(ctx context.Context) (kafka.Message, error) {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return kafka.Message{}, fmt.Errorf("fetch message: %w", err)
	}
	return msg, nil
}

// CommitMessage explicitly commits the offset for the given message.
// Call this only after the message has been successfully persisted.
func (c *Consumer) CommitMessage(ctx context.Context, msg kafka.Message) error {
	if err := c.reader.CommitMessages(ctx, msg); err != nil {
		return fmt.Errorf("commit offset: %w", err)
	}
	return nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

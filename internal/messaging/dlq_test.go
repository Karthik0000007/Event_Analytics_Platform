package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
)

// ──────────────────────────────────────────────────────────────
// DLQ envelope tests
// ──────────────────────────────────────────────────────────────

func TestDLQMessage_Serialization(t *testing.T) {
	original := kafka.Message{
		Topic:     "events",
		Partition: 2,
		Offset:    42,
		Key:       []byte("test-key"),
		Value:     []byte(`{"event_id":"abc","event_type":"click"}`),
	}

	envelope := DLQMessage{
		OriginalTopic:     original.Topic,
		OriginalPartition: original.Partition,
		OriginalOffset:    original.Offset,
		OriginalKey:       string(original.Key),
		OriginalValue:     original.Value,
		ErrorMessage:      "db insert failed",
		ErrorKind:         ErrTransient.String(),
		Retries:           3,
		FailedAt:          time.Now(),
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal DLQ message: %v", err)
	}

	var decoded DLQMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal DLQ message: %v", err)
	}

	if decoded.OriginalTopic != "events" {
		t.Errorf("expected topic 'events', got %q", decoded.OriginalTopic)
	}
	if decoded.OriginalOffset != 42 {
		t.Errorf("expected offset 42, got %d", decoded.OriginalOffset)
	}
	if decoded.ErrorKind != "transient" {
		t.Errorf("expected error_kind 'transient', got %q", decoded.ErrorKind)
	}
	if decoded.Retries != 3 {
		t.Errorf("expected retries 3, got %d", decoded.Retries)
	}
}

// ──────────────────────────────────────────────────────────────
// Failure-injection: simulated DLQ send failures
// ──────────────────────────────────────────────────────────────

func TestDLQProducer_SendWithCancelledContext(t *testing.T) {
	// Sending to a non-existent broker with a cancelled context should fail fast.
	dlq := NewDLQProducer([]string{"localhost:19999"}, "test.dlq")
	defer dlq.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	msg := kafka.Message{
		Key:   []byte("k"),
		Value: []byte("v"),
	}

	err := dlq.Send(ctx, msg, errors.New("test"), ErrPermanent, 1)
	if err == nil {
		t.Error("expected error when sending to DLQ with cancelled context")
	}
}

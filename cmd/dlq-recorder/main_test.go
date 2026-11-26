package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Karthik0000007/Event_Analytics_Platform/internal/messaging"
	"github.com/segmentio/kafka-go"
)

func TestToDeadLetter_ReadsEnvelope(t *testing.T) {
	failedAt := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	env, _ := json.Marshal(messaging.DLQMessage{
		OriginalKey:   "evt-1",
		OriginalValue: `{not json`,
		ErrorKind:     "permanent",
		ErrorMessage:  "[permanent] invalid JSON",
		Retries:       0,
		FailedAt:      failedAt,
	})

	d := toDeadLetter(kafka.Message{Partition: 0, Offset: 42, Value: env})

	if d.OriginalKey != "evt-1" || d.OriginalValue != `{not json` || d.ErrorKind != "permanent" {
		t.Errorf("envelope fields not copied: %+v", d)
	}
	if d.KafkaOffset != 42 || !d.FailedAt.Equal(failedAt) {
		t.Errorf("offset or time wrong: %+v", d)
	}
}

func TestToDeadLetter_KeepsUnreadableEnvelope(t *testing.T) {
	d := toDeadLetter(kafka.Message{Offset: 3, Key: []byte("k"), Value: []byte("garbage"), Time: time.Now()})

	if d.ErrorKind != "unknown" || d.OriginalValue != "garbage" {
		t.Errorf("raw value must be kept when the envelope is unreadable: %+v", d)
	}
	if !strings.HasPrefix(d.ErrorMessage, "unreadable DLQ envelope") {
		t.Errorf("unexpected error message: %q", d.ErrorMessage)
	}
}

func TestToDeadLetter_DefaultsMissingTime(t *testing.T) {
	d := toDeadLetter(kafka.Message{Value: []byte(`{"error_kind":"transient"}`)})
	if d.FailedAt.IsZero() {
		t.Error("failed_at must never be zero")
	}
}

func TestCleanText_RemovesBytesPostgresRejects(t *testing.T) {
	got := cleanText("a\x00b\xffc")
	if strings.ContainsRune(got, 0) || !strings.HasPrefix(got, "a") || !strings.HasSuffix(got, "c") {
		t.Errorf("NUL or invalid UTF-8 not cleaned: %q", got)
	}
}

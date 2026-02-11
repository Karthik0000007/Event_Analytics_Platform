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
// Failure injection scenarios
//
// These tests verify the consumer's decision logic under various
// failure modes WITHOUT requiring a running Kafka / Postgres.
// They exercise classification → retry → DLQ routing paths.
// ──────────────────────────────────────────────────────────────

// --- helpers ---

type insertFunc func(ctx context.Context, id, typ string, payload json.RawMessage) error

// simulateProcessing mirrors the consumer's processMessage logic
// in a unit-testable form. Returns (dlqSent, retries, finalErr).
func simulateProcessing(
	rawValue []byte,
	insertFn insertFunc,
	retryCfg RetryConfig,
) (dlqSent bool, dlqKind ErrorKind, retries int, finalErr error) {
	// Step 1: Unmarshal (poison pill check)
	var evt struct {
		EventID   string          `json:"event_id"`
		EventType string          `json:"event_type"`
		Payload   json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(rawValue, &evt); err != nil {
		return true, ErrPermanent, 0, err
	}

	// Step 2: Validate required fields
	if evt.EventID == "" || evt.EventType == "" {
		return true, ErrPermanent, 0, NewPermanent("missing required fields", nil)
	}

	// Step 3: Bounded retry loop
	for attempt := 0; ; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		err := insertFn(ctx, evt.EventID, evt.EventType, evt.Payload)
		cancel()

		if err == nil {
			return false, 0, attempt, nil
		}

		kind := Classify(err)

		if kind == ErrPermanent {
			return true, kind, attempt + 1, err
		}
		if !retryCfg.ShouldRetry(kind, attempt) {
			return true, kind, attempt + 1, err
		}
		// (skip sleep in tests)
	}
}

// ── Scenario 1: Poison pill (invalid JSON) ────────────────────

func TestFailureInjection_PoisonPill_InvalidJSON(t *testing.T) {
	raw := []byte(`{not-valid-json!!!}`)
	insert := func(ctx context.Context, id, typ string, payload json.RawMessage) error {
		t.Fatal("insert should never be called for poison pill")
		return nil
	}

	dlqSent, kind, retries, err := simulateProcessing(raw, insert, DefaultRetryConfig())

	if !dlqSent {
		t.Error("expected poison pill to be routed to DLQ")
	}
	if kind != ErrPermanent {
		t.Errorf("expected permanent error, got %v", kind)
	}
	if retries != 0 {
		t.Errorf("expected 0 retries for poison pill, got %d", retries)
	}
	if err == nil {
		t.Error("expected non-nil error")
	}
}

// ── Scenario 2: Poison pill (missing required fields) ─────────

func TestFailureInjection_PoisonPill_MissingFields(t *testing.T) {
	raw := []byte(`{"event_id":"","event_type":"click","payload":{}}`)
	insert := func(ctx context.Context, id, typ string, payload json.RawMessage) error {
		t.Fatal("insert should never be called for missing fields")
		return nil
	}

	dlqSent, kind, _, _ := simulateProcessing(raw, insert, DefaultRetryConfig())

	if !dlqSent {
		t.Error("expected missing-field message to be routed to DLQ")
	}
	if kind != ErrPermanent {
		t.Error("expected permanent classification for missing fields")
	}
}

// ── Scenario 3: Transient failure recovers within budget ──────

func TestFailureInjection_TransientRecovery(t *testing.T) {
	callCount := 0
	insert := func(ctx context.Context, id, typ string, payload json.RawMessage) error {
		callCount++
		if callCount < 3 {
			return errors.New("connection refused") // transient
		}
		return nil // succeeds on 3rd call
	}

	raw := []byte(`{"event_id":"e1","event_type":"click","payload":{}}`)
	rc := DefaultRetryConfig()
	rc.MaxRetries = 5

	dlqSent, _, retries, err := simulateProcessing(raw, insert, rc)

	if dlqSent {
		t.Error("expected successful processing, not DLQ")
	}
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if retries != 2 { // 0-indexed: attempts 0, 1 failed; 2 succeeded
		t.Errorf("expected 2 retries before success, got %d", retries)
	}
}

// ── Scenario 4: Transient failure exhausts retry budget ───────

func TestFailureInjection_TransientExhaustsRetries(t *testing.T) {
	insert := func(ctx context.Context, id, typ string, payload json.RawMessage) error {
		return errors.New("connection reset by peer") // always transient
	}

	raw := []byte(`{"event_id":"e2","event_type":"view","payload":{}}`)
	rc := DefaultRetryConfig()
	rc.MaxRetries = 3

	dlqSent, kind, retries, err := simulateProcessing(raw, insert, rc)

	if !dlqSent {
		t.Error("expected DLQ routing after exhausting retries")
	}
	if kind != ErrTransient {
		t.Errorf("expected transient classification, got %v", kind)
	}
	if retries != 4 {
		t.Errorf("expected 4 attempts (3 retries + 1 final), got %d", retries)
	}
	if err == nil {
		t.Error("expected non-nil error")
	}
}

// ── Scenario 5: Permanent failure (constraint violation) ──────

func TestFailureInjection_PermanentConstraintViolation(t *testing.T) {
	insert := func(ctx context.Context, id, typ string, payload json.RawMessage) error {
		return errors.New(`pq: value too long violates check constraint "events_type_len"`)
	}

	raw := []byte(`{"event_id":"e3","event_type":"x","payload":{}}`)
	rc := DefaultRetryConfig()
	rc.MaxRetries = 5

	dlqSent, kind, retries, _ := simulateProcessing(raw, insert, rc)

	if !dlqSent {
		t.Error("expected DLQ for permanent error")
	}
	if kind != ErrPermanent {
		t.Error("expected permanent classification")
	}
	if retries != 1 {
		t.Errorf("expected exactly 1 attempt (no retries), got %d", retries)
	}
}

// ── Scenario 6: First call succeeds (happy path) ─────────────

func TestFailureInjection_HappyPath(t *testing.T) {
	insert := func(ctx context.Context, id, typ string, payload json.RawMessage) error {
		return nil
	}

	raw := []byte(`{"event_id":"e4","event_type":"purchase","payload":{"amount":99}}`)
	dlqSent, _, retries, err := simulateProcessing(raw, insert, DefaultRetryConfig())

	if dlqSent {
		t.Error("expected no DLQ for successful processing")
	}
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if retries != 0 {
		t.Errorf("expected 0 retries for happy path, got %d", retries)
	}
}

// ── Scenario 7: Verify DLQ envelope contains full context ────

func TestFailureInjection_DLQEnvelopeContainsContext(t *testing.T) {
	original := kafka.Message{
		Topic:     "events",
		Partition: 1,
		Offset:    100,
		Key:       []byte("user-123"),
		Value:     []byte(`{"event_id":"e5","event_type":"click"}`),
	}

	reason := NewPermanent("constraint violation", errors.New("pq: unique constraint"))
	envelope := DLQMessage{
		OriginalTopic:     original.Topic,
		OriginalPartition: original.Partition,
		OriginalOffset:    original.Offset,
		OriginalKey:       string(original.Key),
		OriginalValue:     original.Value,
		ErrorMessage:      reason.Error(),
		ErrorKind:         reason.Kind.String(),
		Retries:           2,
		FailedAt:          time.Now(),
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded DLQMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Verify all context is preserved
	if decoded.OriginalTopic != "events" {
		t.Error("original topic lost")
	}
	if decoded.OriginalPartition != 1 {
		t.Error("original partition lost")
	}
	if decoded.OriginalOffset != 100 {
		t.Error("original offset lost")
	}
	if decoded.OriginalKey != "user-123" {
		t.Error("original key lost")
	}
	if decoded.ErrorKind != "permanent" {
		t.Error("error kind lost")
	}
	if decoded.Retries != 2 {
		t.Error("retry count lost")
	}

	// Verify original event can still be deserialized
	var evt struct {
		EventID   string `json:"event_id"`
		EventType string `json:"event_type"`
	}
	if err := json.Unmarshal(decoded.OriginalValue, &evt); err != nil {
		t.Fatal("could not deserialize original event from DLQ envelope")
	}
	if evt.EventID != "e5" || evt.EventType != "click" {
		t.Error("original event data corrupted in DLQ envelope")
	}
}

// ── Scenario 8: Zero-retry config sends to DLQ immediately ───

func TestFailureInjection_ZeroRetriesGoStraightToDLQ(t *testing.T) {
	insert := func(ctx context.Context, id, typ string, payload json.RawMessage) error {
		return errors.New("connection timeout")
	}

	raw := []byte(`{"event_id":"e6","event_type":"view","payload":{}}`)
	rc := RetryConfig{
		MaxRetries: 0, // no retries allowed
		BaseDelay:  time.Millisecond,
		MaxDelay:   time.Millisecond,
		Multiplier: 1,
	}

	dlqSent, _, retries, _ := simulateProcessing(raw, insert, rc)

	if !dlqSent {
		t.Error("expected DLQ with zero-retry config")
	}
	if retries != 1 {
		t.Errorf("expected 1 attempt, got %d", retries)
	}
}

package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/Karthik0000007/Event_Analytics_Platform/internal/logging"
	"github.com/segmentio/kafka-go"
)

// ──────────────────────────────────────────────────────────────
// Failure injection scenarios
//
// These tests run the real Processor (the code the consumer binary uses)
// against fakes for Postgres, the DLQ and the offset committer, so they
// need no running Kafka or Postgres.
// ──────────────────────────────────────────────────────────────

// fakeStore fails inserts and pings according to the functions it is given.
type fakeStore struct {
	insertErr func(call int) error // nil means every insert succeeds
	pingErr   func(call int) error // nil means every ping succeeds
	inserts   int
	pings     int
	stored    []string
}

func (f *fakeStore) InsertEvent(_ context.Context, id, _ string, _ json.RawMessage) error {
	f.inserts++
	if f.insertErr != nil {
		if err := f.insertErr(f.inserts); err != nil {
			return err
		}
	}
	f.stored = append(f.stored, id)
	return nil
}

func (f *fakeStore) Ping(context.Context) error {
	f.pings++
	if f.pingErr != nil {
		return f.pingErr(f.pings)
	}
	return nil
}

// fakeDLQ records dead letters. The first failFirst sends return an error.
type fakeDLQ struct {
	failFirst int
	attempts  int
	kinds     []ErrorKind
	retries   []int
}

func (f *fakeDLQ) Send(_ context.Context, _ kafka.Message, _ error, kind ErrorKind, retries int) error {
	f.attempts++
	if f.attempts <= f.failFirst {
		return errors.New("kafka: broker not available")
	}
	f.kinds = append(f.kinds, kind)
	f.retries = append(f.retries, retries)
	return nil
}

// fakeOffsets records which offsets were committed.
type fakeOffsets struct{ committed []int64 }

func (f *fakeOffsets) CommitMessage(_ context.Context, msg kafka.Message) error {
	f.committed = append(f.committed, msg.Offset)
	return nil
}

// fastRetry keeps tests quick: tiny delays, MaxRetries retries.
func fastRetry(maxRetries int) RetryConfig {
	return RetryConfig{MaxRetries: maxRetries, BaseDelay: time.Millisecond, MaxDelay: 2 * time.Millisecond, Multiplier: 2}
}

type harness struct {
	store   *fakeStore
	dlq     *fakeDLQ
	offsets *fakeOffsets
	proc    *Processor
}

func newHarness(store *fakeStore, dlq *fakeDLQ, maxRetries int) *harness {
	offsets := &fakeOffsets{}
	return &harness{
		store: store, dlq: dlq, offsets: offsets,
		proc: &Processor{
			Store: store, DLQ: dlq, Offsets: offsets,
			Retry:      fastRetry(maxRetries),
			Logger:     logging.NewWriter("test", io.Discard),
			OutagePoll: time.Millisecond,
		},
	}
}

func message(offset int64, value string) kafka.Message {
	return kafka.Message{Topic: "events", Offset: offset, Value: []byte(value)}
}

const goodEvent = `{"event_id":"e1","event_type":"click","payload":{}}`

// ── Scenario 1: Poison pill (invalid JSON) ────────────────────

func TestFailureInjection_PoisonPill_InvalidJSON(t *testing.T) {
	h := newHarness(&fakeStore{}, &fakeDLQ{}, 5)

	if err := h.proc.Process(context.Background(), message(7, `{not-valid-json!!!}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.store.inserts != 0 {
		t.Error("insert must never be called for a poison pill")
	}
	if len(h.dlq.kinds) != 1 || h.dlq.kinds[0] != ErrPermanent || h.dlq.retries[0] != 0 {
		t.Errorf("expected one permanent dead letter with 0 retries, got kinds=%v retries=%v", h.dlq.kinds, h.dlq.retries)
	}
	if len(h.offsets.committed) != 1 || h.offsets.committed[0] != 7 {
		t.Errorf("expected offset 7 committed after DLQ, got %v", h.offsets.committed)
	}
}

// ── Scenario 2: Poison pill (missing required fields) ─────────

func TestFailureInjection_PoisonPill_MissingFields(t *testing.T) {
	h := newHarness(&fakeStore{}, &fakeDLQ{}, 5)

	_ = h.proc.Process(context.Background(), message(1, `{"event_id":"","event_type":"click","payload":{}}`))

	if h.store.inserts != 0 || len(h.dlq.kinds) != 1 || h.dlq.kinds[0] != ErrPermanent {
		t.Errorf("expected permanent DLQ without insert, got inserts=%d kinds=%v", h.store.inserts, h.dlq.kinds)
	}
}

// ── Scenario 3: Transient failure recovers within budget ──────

func TestFailureInjection_TransientRecovery(t *testing.T) {
	store := &fakeStore{insertErr: func(call int) error {
		if call < 3 {
			return errors.New("connection refused")
		}
		return nil
	}}
	h := newHarness(store, &fakeDLQ{}, 5)

	_ = h.proc.Process(context.Background(), message(2, goodEvent))

	if store.inserts != 3 || len(store.stored) != 1 {
		t.Errorf("expected success on the 3rd insert, got inserts=%d stored=%v", store.inserts, store.stored)
	}
	if len(h.dlq.kinds) != 0 {
		t.Error("a recovered transient error must not reach the DLQ")
	}
	if len(h.offsets.committed) != 1 {
		t.Errorf("expected exactly one commit, got %v", h.offsets.committed)
	}
}

// ── Scenario 4: Transient failure exhausts budget while the DB is healthy ──

func TestFailureInjection_TransientExhaustsRetries(t *testing.T) {
	store := &fakeStore{insertErr: func(int) error { return errors.New("connection reset by peer") }}
	h := newHarness(store, &fakeDLQ{}, 3)

	_ = h.proc.Process(context.Background(), message(3, goodEvent))

	if store.inserts != 4 {
		t.Errorf("expected 4 attempts (1 + 3 retries), got %d", store.inserts)
	}
	if len(h.dlq.kinds) != 1 || h.dlq.kinds[0] != ErrTransient || h.dlq.retries[0] != 4 {
		t.Errorf("expected one transient dead letter after 4 attempts, got kinds=%v retries=%v", h.dlq.kinds, h.dlq.retries)
	}
}

// ── Scenario 5: Permanent failure (constraint violation) ──────

func TestFailureInjection_PermanentConstraintViolation(t *testing.T) {
	store := &fakeStore{insertErr: func(int) error {
		return errors.New(`pq: value too long violates check constraint "events_type_len"`)
	}}
	h := newHarness(store, &fakeDLQ{}, 5)

	_ = h.proc.Process(context.Background(), message(4, goodEvent))

	if store.inserts != 1 {
		t.Errorf("expected exactly 1 attempt (no retries), got %d", store.inserts)
	}
	if len(h.dlq.kinds) != 1 || h.dlq.kinds[0] != ErrPermanent {
		t.Errorf("expected a permanent dead letter, got %v", h.dlq.kinds)
	}
}

// ── Scenario 6: First call succeeds (happy path) ─────────────

func TestFailureInjection_HappyPath(t *testing.T) {
	h := newHarness(&fakeStore{}, &fakeDLQ{}, 5)

	_ = h.proc.Process(context.Background(), message(5, goodEvent))

	if h.store.inserts != 1 || len(h.dlq.kinds) != 0 {
		t.Errorf("expected one insert and no DLQ, got inserts=%d dlq=%v", h.store.inserts, h.dlq.kinds)
	}
	if len(h.offsets.committed) != 1 || h.offsets.committed[0] != 5 {
		t.Errorf("expected offset 5 committed, got %v", h.offsets.committed)
	}
}

// ── Scenario 7: DLQ envelope keeps the full context ───────────

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
		OriginalValue:     string(original.Value),
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

	if decoded.OriginalTopic != "events" || decoded.OriginalPartition != 1 || decoded.OriginalOffset != 100 {
		t.Error("original topic, partition or offset lost")
	}
	if decoded.OriginalKey != "user-123" || decoded.ErrorKind != "permanent" || decoded.Retries != 2 {
		t.Error("key, error kind or retry count lost")
	}

	var evt struct {
		EventID   string `json:"event_id"`
		EventType string `json:"event_type"`
	}
	if err := json.Unmarshal([]byte(decoded.OriginalValue), &evt); err != nil {
		t.Fatal("could not deserialize original event from DLQ envelope")
	}
	if evt.EventID != "e5" || evt.EventType != "click" {
		t.Error("original event data corrupted in DLQ envelope")
	}
}

// ── Scenario 8: Zero-retry config sends to DLQ after one attempt ──

func TestFailureInjection_ZeroRetriesGoStraightToDLQ(t *testing.T) {
	store := &fakeStore{insertErr: func(int) error { return errors.New("connection timeout") }}
	h := newHarness(store, &fakeDLQ{}, 0)

	_ = h.proc.Process(context.Background(), message(6, goodEvent))

	if store.inserts != 1 || len(h.dlq.kinds) != 1 {
		t.Errorf("expected 1 attempt then DLQ, got inserts=%d dlq=%v", store.inserts, h.dlq.kinds)
	}
}

// ── Scenario 9: Database outage pauses instead of dead-lettering ──

func TestFailureInjection_DatabaseOutagePausesThenRecovers(t *testing.T) {
	const outageInserts = 4 // inserts fail until the database "comes back"
	store := &fakeStore{
		insertErr: func(call int) error {
			if call <= outageInserts {
				return errors.New("dial tcp: lookup postgres: no such host")
			}
			return nil
		},
		// The first two pings fail (database down), then it answers again.
		pingErr: func(call int) error {
			if call <= 2 {
				return errors.New("connection refused")
			}
			return nil
		},
	}
	h := newHarness(store, &fakeDLQ{}, 2)

	if err := h.proc.Process(context.Background(), message(8, goodEvent)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(h.dlq.kinds) != 0 {
		t.Fatalf("an outage must not send events to the DLQ, got %v", h.dlq.kinds)
	}
	if len(store.stored) != 1 || len(h.offsets.committed) != 1 {
		t.Errorf("expected the event stored and committed once, got stored=%v committed=%v", store.stored, h.offsets.committed)
	}
}

// ── Scenario 10: DLQ write fails, offset waits until it succeeds ──

func TestFailureInjection_DLQWriteFailureDoesNotCommitEarly(t *testing.T) {
	dlq := &fakeDLQ{failFirst: 2}
	h := newHarness(&fakeStore{}, dlq, 5)

	_ = h.proc.Process(context.Background(), message(9, `not json`))

	if dlq.attempts != 3 || len(dlq.kinds) != 1 {
		t.Errorf("expected 2 failed DLQ writes then 1 success, got attempts=%d sent=%d", dlq.attempts, len(dlq.kinds))
	}
	if len(h.offsets.committed) != 1 {
		t.Errorf("expected a single commit after the DLQ write succeeded, got %v", h.offsets.committed)
	}
}

// ── Scenario 11: Shutdown during retries leaves the offset uncommitted ──

func TestFailureInjection_ShutdownDuringRetryDoesNotCommit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	store := &fakeStore{insertErr: func(int) error {
		cancel() // simulate SIGTERM arriving while the insert is failing
		return errors.New("connection refused")
	}}
	h := newHarness(store, &fakeDLQ{}, 5)

	err := h.proc.Process(ctx, message(10, goodEvent))

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if len(h.offsets.committed) != 0 || len(h.dlq.kinds) != 0 {
		t.Errorf("shutdown must not commit or dead-letter, got commits=%v dlq=%v", h.offsets.committed, h.dlq.kinds)
	}
}

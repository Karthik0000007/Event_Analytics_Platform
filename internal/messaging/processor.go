package messaging

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Karthik0000007/Event_Analytics_Platform/internal/logging"
	"github.com/segmentio/kafka-go"
)

// EventStore persists events. *storage.DB implements it.
type EventStore interface {
	InsertEvent(ctx context.Context, eventID, eventType string, payload json.RawMessage) error
	Ping(ctx context.Context) error
}

// DeadLetterWriter sends a failed message to the dead-letter topic. *DLQProducer implements it.
type DeadLetterWriter interface {
	Send(ctx context.Context, original kafka.Message, reason error, kind ErrorKind, retries int) error
}

// OffsetCommitter commits a Kafka offset. *Consumer implements it.
type OffsetCommitter interface {
	CommitMessage(ctx context.Context, msg kafka.Message) error
}

// Event is the message value the consumer expects on the events topic.
type Event struct {
	EventID   string          `json:"event_id"`
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
}

// Processor decides what happens to one Kafka message:
// store it, retry it, pause during a database outage, or dead-letter it.
// The offset is committed only after the event is in Postgres or in the DLQ.
type Processor struct {
	Store   EventStore
	DLQ     DeadLetterWriter
	Offsets OffsetCommitter
	Retry   RetryConfig
	Logger  *logging.Logger

	// OutagePoll is how often the database is pinged while it is down (default 2s).
	OutagePoll time.Duration
	// InsertTimeout bounds one insert attempt (default 5s).
	InsertTimeout time.Duration
}

// Process handles one message. It returns an error only when ctx is cancelled
// before the message is finished. The offset is then left uncommitted, so Kafka
// delivers the message again after a restart.
func (p *Processor) Process(ctx context.Context, msg kafka.Message) error {
	evt, err := DecodeEvent(msg.Value)
	if err != nil {
		p.Logger.Error("poison message, sending to DLQ", map[string]any{
			"error": err.Error(), "offset": msg.Offset, "partition": msg.Partition,
		})
		return p.deadLetter(ctx, msg, err, ErrPermanent, 0)
	}
	return p.store(ctx, msg, evt)
}

// DecodeEvent parses a message value. Any error it returns is permanent,
// because retrying cannot fix bad JSON or a missing field.
func DecodeEvent(value []byte) (Event, error) {
	var evt Event
	if err := json.Unmarshal(value, &evt); err != nil {
		return evt, NewPermanent("invalid JSON", err)
	}
	if evt.EventID == "" || evt.EventType == "" {
		return evt, NewPermanent("missing event_id or event_type", nil)
	}
	return evt, nil
}

// store inserts the event with bounded retries.
func (p *Processor) store(ctx context.Context, msg kafka.Message, evt Event) error {
	for attempt := 0; ; attempt++ {
		err := p.insert(ctx, evt)
		if err == nil {
			p.commit(ctx, msg)
			p.Logger.Info("event persisted", map[string]any{
				"event_id": evt.EventID, "offset": msg.Offset, "attempts": attempt + 1,
			})
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}

		kind := Classify(err)
		p.Logger.Error("insert failed", map[string]any{
			"event_id": evt.EventID, "error": err.Error(), "error_kind": kind.String(), "attempt": attempt + 1,
		})

		switch {
		case kind == ErrPermanent:
			return p.deadLetter(ctx, msg, err, kind, attempt+1)
		case p.Retry.ShouldRetry(kind, attempt):
			if err := p.Retry.Sleep(ctx, attempt); err != nil {
				return err
			}
		case p.storeIsDown(ctx):
			// The database itself is down, so this message is not the problem.
			// Wait for recovery, then try again with a fresh retry budget.
			if err := p.waitForStore(ctx); err != nil {
				return err
			}
			attempt = -1
		default:
			return p.deadLetter(ctx, msg, err, kind, attempt+1)
		}
	}
}

func (p *Processor) insert(ctx context.Context, evt Event) error {
	timeout := p.InsertTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ictx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return p.Store.InsertEvent(ictx, evt.EventID, evt.EventType, evt.Payload)
}

// deadLetter writes the message to the DLQ, retrying until Kafka accepts it.
// The offset is committed only after the DLQ write succeeds, so a failed
// DLQ write never loses the message.
func (p *Processor) deadLetter(ctx context.Context, msg kafka.Message, reason error, kind ErrorKind, retries int) error {
	for attempt := 0; ; attempt++ {
		err := p.DLQ.Send(ctx, msg, reason, kind, retries)
		if err == nil {
			p.commit(ctx, msg)
			p.Logger.Info("message sent to DLQ", map[string]any{
				"offset": msg.Offset, "error_kind": kind.String(), "retries": retries,
			})
			return nil
		}
		p.Logger.Error("DLQ write failed, retrying; offset not committed", map[string]any{
			"offset": msg.Offset, "error": err.Error(), "attempt": attempt + 1,
		})
		if err := p.Retry.Sleep(ctx, attempt); err != nil {
			return err
		}
	}
}

// commit logs a failed commit but does not stop. The next successful commit
// covers this offset, and a replay is harmless because inserts are idempotent.
func (p *Processor) commit(ctx context.Context, msg kafka.Message) {
	if err := p.Offsets.CommitMessage(ctx, msg); err != nil {
		p.Logger.Error("offset commit failed", map[string]any{"offset": msg.Offset, "error": err.Error()})
	}
}

func (p *Processor) storeIsDown(ctx context.Context) bool {
	pctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return p.Store.Ping(pctx) != nil
}

// waitForStore blocks until the database answers a ping or ctx is cancelled.
func (p *Processor) waitForStore(ctx context.Context) error {
	poll := p.OutagePoll
	if poll <= 0 {
		poll = 2 * time.Second
	}
	p.Logger.Error("database unreachable, pausing until it recovers", nil)

	ticker := time.NewTicker(poll)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
		if !p.storeIsDown(ctx) {
			p.Logger.Info("database is back, resuming", nil)
			return nil
		}
	}
}

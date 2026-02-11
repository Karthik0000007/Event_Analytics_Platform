package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Karthik0000007/Event_Analytics_Platform/internal/config"
	"github.com/Karthik0000007/Event_Analytics_Platform/internal/logging"
	"github.com/Karthik0000007/Event_Analytics_Platform/internal/messaging"
	"github.com/Karthik0000007/Event_Analytics_Platform/internal/storage"
	"github.com/segmentio/kafka-go"
)

type event struct {
	EventID   string          `json:"event_id"`
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
}

func main() {
	cfg := config.Load()
	cfg.ServiceName = "event-consumer"
	logger := logging.New(cfg.ServiceName)

	// ── Retry config ───────────────────────────────────────────
	retryCfg := messaging.DefaultRetryConfig()
	retryCfg.MaxRetries = cfg.MaxRetries

	// ── Connect to PostgreSQL ──────────────────────────────────
	db, err := storage.New(cfg.DatabaseDSN)
	if err != nil {
		logger.Error("failed to connect to postgres", map[string]any{"error": err.Error()})
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("connected to postgres", map[string]any{})

	// ── Create Kafka consumer ──────────────────────────────────
	consumer := messaging.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID)
	defer consumer.Close()
	logger.Info("kafka consumer started", map[string]any{
		"brokers": cfg.KafkaBrokers,
		"topic":   cfg.KafkaTopic,
		"group":   cfg.KafkaGroupID,
	})

	// ── Create DLQ producer ────────────────────────────────────
	dlq := messaging.NewDLQProducer(cfg.KafkaBrokers, cfg.KafkaDLQTopic)
	defer dlq.Close()
	logger.Info("DLQ producer ready", map[string]any{"dlq_topic": cfg.KafkaDLQTopic})

	// ── Graceful shutdown ──────────────────────────────────────
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		logger.Info("shutdown signal received", map[string]any{"signal": sig.String()})
		cancel()
	}()

	// ── Consume loop ───────────────────────────────────────────
	logger.Info("consuming events", map[string]any{})

	for {
		// 1. Fetch message (blocks until available or context cancelled)
		msg, err := consumer.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				logger.Info("consumer shutting down", map[string]any{})
				return
			}
			logger.Error("fetch failed", map[string]any{"error": err.Error()})
			continue
		}

		// 2. Process with retry + DLQ routing
		processMessage(ctx, logger, db, consumer, dlq, retryCfg, msg)
	}
}

// processMessage handles deserialization, persistence, retry, and DLQ routing.
func processMessage(
	ctx context.Context,
	logger *logging.Logger,
	db *storage.DB,
	consumer *messaging.Consumer,
	dlq *messaging.DLQProducer,
	retryCfg messaging.RetryConfig,
	msg kafka.Message,
) {
	// ── Poison pill check: unmarshal ───────────────────────────
	var evt event
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		logger.Error("poison pill: invalid JSON, routing to DLQ", map[string]any{
			"error":     err.Error(),
			"offset":    msg.Offset,
			"partition": msg.Partition,
			"raw_size":  len(msg.Value),
		})
		sendToDLQ(ctx, logger, dlq, msg, err, messaging.ErrPermanent, 0)
		commitAndLog(ctx, logger, consumer, msg, "poison-pill")
		return
	}

	// ── Poison pill check: required fields ─────────────────────
	if evt.EventID == "" || evt.EventType == "" {
		err := messaging.NewPermanent("missing required fields", nil)
		logger.Error("poison pill: missing event_id or event_type", map[string]any{
			"offset":    msg.Offset,
			"partition": msg.Partition,
		})
		sendToDLQ(ctx, logger, dlq, msg, err, messaging.ErrPermanent, 0)
		commitAndLog(ctx, logger, consumer, msg, "poison-pill-missing-fields")
		return
	}

	// ── Bounded retry loop for DB insert ──────────────────────
	var lastErr error
	for attempt := 0; ; attempt++ {
		dbCtx, dbCancel := context.WithTimeout(ctx, 5*time.Second)
		lastErr = db.InsertEvent(dbCtx, evt.EventID, evt.EventType, evt.Payload)
		dbCancel()

		if lastErr == nil {
			// Success — commit offset
			if err := consumer.CommitMessage(ctx, msg); err != nil {
				logger.Error("offset commit failed (event already in DB)", map[string]any{
					"event_id": evt.EventID,
					"error":    err.Error(),
				})
			}
			logger.Info("event persisted", map[string]any{
				"event_id":   evt.EventID,
				"event_type": evt.EventType,
				"offset":     msg.Offset,
				"attempts":   attempt + 1,
			})
			return
		}

		kind := messaging.Classify(lastErr)
		logger.Error("db insert failed", map[string]any{
			"event_id":   evt.EventID,
			"error":      lastErr.Error(),
			"error_kind": kind.String(),
			"attempt":    attempt + 1,
			"max":        retryCfg.MaxRetries,
		})

		// Permanent error — no point retrying
		if kind == messaging.ErrPermanent {
			sendToDLQ(ctx, logger, dlq, msg, lastErr, kind, attempt+1)
			commitAndLog(ctx, logger, consumer, msg, "permanent-error")
			return
		}

		// Transient but budget exhausted
		if !retryCfg.ShouldRetry(kind, attempt) {
			logger.Error("retry budget exhausted, routing to DLQ", map[string]any{
				"event_id": evt.EventID,
				"retries":  attempt + 1,
			})
			sendToDLQ(ctx, logger, dlq, msg, lastErr, kind, attempt+1)
			commitAndLog(ctx, logger, consumer, msg, "retries-exhausted")
			return
		}

		// Back off before next attempt
		if err := retryCfg.Sleep(ctx, attempt); err != nil {
			// Context cancelled during sleep — exit without commit
			logger.Info("retry sleep interrupted by shutdown", map[string]any{
				"event_id": evt.EventID,
			})
			return
		}
	}
}

// sendToDLQ routes a message to the dead-letter topic, logging failures.
func sendToDLQ(
	ctx context.Context,
	logger *logging.Logger,
	dlq *messaging.DLQProducer,
	msg kafka.Message,
	reason error,
	kind messaging.ErrorKind,
	retries int,
) {
	if err := dlq.Send(ctx, msg, reason, kind, retries); err != nil {
		logger.Error("CRITICAL: failed to write to DLQ", map[string]any{
			"error":          err.Error(),
			"original_error": reason.Error(),
			"offset":         msg.Offset,
		})
	} else {
		logger.Info("message routed to DLQ", map[string]any{
			"offset":     msg.Offset,
			"error_kind": kind.String(),
			"retries":    retries,
		})
	}
}

// commitAndLog commits the offset so the consumer moves past the failed message.
func commitAndLog(
	ctx context.Context,
	logger *logging.Logger,
	consumer *messaging.Consumer,
	msg kafka.Message,
	reason string,
) {
	if err := consumer.CommitMessage(ctx, msg); err != nil {
		logger.Error("offset commit failed after DLQ routing", map[string]any{
			"error":  err.Error(),
			"reason": reason,
			"offset": msg.Offset,
		})
	}
}

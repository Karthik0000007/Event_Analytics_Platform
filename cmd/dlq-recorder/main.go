// Command dlq-recorder reads the dead-letter topic (events.dlq) and stores
// each message in the dead_letters table, so the API and dashboard can show it.
//
// The offset is committed only after the insert succeeds. If Postgres is down,
// the recorder keeps retrying and the dead letters wait safely in Kafka.
package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Karthik0000007/Event_Analytics_Platform/internal/config"
	"github.com/Karthik0000007/Event_Analytics_Platform/internal/logging"
	"github.com/Karthik0000007/Event_Analytics_Platform/internal/messaging"
	"github.com/Karthik0000007/Event_Analytics_Platform/internal/storage"
	"github.com/segmentio/kafka-go"
)

func main() {
	cfg := config.Load()
	logger := logging.New("dlq-recorder")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := storage.New(cfg.DatabaseDSN)
	if err != nil {
		logger.Error("failed to connect to postgres", map[string]any{"error": err.Error()})
		os.Exit(1)
	}
	defer db.Close()

	consumer := messaging.NewConsumer(cfg.KafkaBrokers, cfg.KafkaDLQTopic, cfg.KafkaGroupID)
	defer consumer.Close()

	logger.Info("recording dead letters", map[string]any{"topic": cfg.KafkaDLQTopic, "group": cfg.KafkaGroupID})
	retry := messaging.DefaultRetryConfig()

	for {
		msg, err := consumer.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			logger.Error("fetch failed", map[string]any{"error": err.Error()})
			_ = retry.Sleep(ctx, 2) // short pause so a broken broker does not cause a busy loop
			continue
		}

		if err := record(ctx, logger, db, retry, toDeadLetter(msg)); err != nil {
			break // shutdown: offset not committed, message is read again on restart
		}
		if err := consumer.CommitMessage(ctx, msg); err != nil {
			logger.Error("offset commit failed", map[string]any{"offset": msg.Offset, "error": err.Error()})
		}
	}
	logger.Info("recorder stopped cleanly", nil)
}

// record inserts a dead letter, retrying transient errors until it works.
// There is no "DLQ for the DLQ": a permanent error is logged in full and skipped.
func record(ctx context.Context, logger *logging.Logger, db *storage.DB, retry messaging.RetryConfig, d storage.DeadLetter) error {
	for attempt := 0; ; attempt++ {
		ictx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := db.InsertDeadLetter(ictx, d)
		cancel()
		if err == nil {
			logger.Info("dead letter recorded", map[string]any{"offset": d.KafkaOffset, "error_kind": d.ErrorKind})
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if messaging.Classify(err) == messaging.ErrPermanent {
			logger.Error("dead letter rejected by postgres, skipping", map[string]any{
				"offset": d.KafkaOffset, "error": err.Error(), "original_value": d.OriginalValue,
			})
			return nil
		}
		logger.Error("insert failed, retrying", map[string]any{"offset": d.KafkaOffset, "error": err.Error(), "attempt": attempt + 1})
		if err := retry.Sleep(ctx, attempt); err != nil {
			return err
		}
	}
}

// toDeadLetter turns a message from events.dlq into a table row.
// If the envelope cannot be parsed, the raw bytes are kept so nothing is lost.
func toDeadLetter(msg kafka.Message) storage.DeadLetter {
	d := storage.DeadLetter{KafkaPartition: msg.Partition, KafkaOffset: msg.Offset}

	var env messaging.DLQMessage
	if err := json.Unmarshal(msg.Value, &env); err != nil {
		d.OriginalKey = cleanText(string(msg.Key))
		d.OriginalValue = cleanText(string(msg.Value))
		d.ErrorKind = "unknown"
		d.ErrorMessage = "unreadable DLQ envelope: " + err.Error()
		d.FailedAt = msg.Time
	} else {
		d.OriginalKey = cleanText(env.OriginalKey)
		d.OriginalValue = cleanText(env.OriginalValue)
		d.ErrorKind = env.ErrorKind
		d.ErrorMessage = cleanText(env.ErrorMessage)
		d.RetryCount = env.Retries
		d.FailedAt = env.FailedAt
	}
	if d.FailedAt.IsZero() {
		d.FailedAt = time.Now()
	}
	return d
}

// cleanText makes a string safe for a Postgres TEXT column, which rejects
// invalid UTF-8 and the NUL character. Poison pills often contain both.
func cleanText(s string) string {
	s = strings.ToValidUTF8(s, "�")
	return strings.ReplaceAll(s, "\x00", "�")
}

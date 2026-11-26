// Command event-consumer reads events from Kafka and stores them in PostgreSQL.
// Failed messages go to the dead-letter topic. See messaging.Processor for the rules.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Karthik0000007/Event_Analytics_Platform/internal/config"
	"github.com/Karthik0000007/Event_Analytics_Platform/internal/logging"
	"github.com/Karthik0000007/Event_Analytics_Platform/internal/messaging"
	"github.com/Karthik0000007/Event_Analytics_Platform/internal/storage"
)

func main() {
	cfg := config.Load()
	logger := logging.New("event-consumer")

	// ctx is cancelled on Ctrl+C (SIGINT) or `docker stop` (SIGTERM).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := storage.New(cfg.DatabaseDSN)
	if err != nil {
		logger.Error("failed to connect to postgres", map[string]any{"error": err.Error()})
		os.Exit(1)
	}
	defer db.Close()

	consumer := messaging.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID)
	defer consumer.Close()

	dlq := messaging.NewDLQProducer(cfg.KafkaBrokers, cfg.KafkaDLQTopic)
	defer dlq.Close()

	retry := messaging.DefaultRetryConfig()
	retry.MaxRetries = cfg.MaxRetries

	proc := &messaging.Processor{Store: db, DLQ: dlq, Offsets: consumer, Retry: retry, Logger: logger}

	logger.Info("consuming events", map[string]any{
		"brokers": cfg.KafkaBrokers, "topic": cfg.KafkaTopic,
		"group": cfg.KafkaGroupID, "dlq_topic": cfg.KafkaDLQTopic,
	})
	consume(ctx, logger, consumer, proc)
	logger.Info("consumer stopped cleanly", nil)
}

// consume fetches one message at a time and hands it to the processor.
// Messages are processed in order, so a partition never skips ahead.
func consume(ctx context.Context, logger *logging.Logger, consumer *messaging.Consumer, proc *messaging.Processor) {
	for {
		msg, err := consumer.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logger.Error("fetch failed", map[string]any{"error": err.Error()})
			sleep(ctx, time.Second)
			continue
		}

		// Process returns an error only when shutdown interrupts it.
		// The offset stays uncommitted, so the message is redelivered on restart.
		if err := proc.Process(ctx, msg); err != nil {
			logger.Info("shutdown interrupted processing, message will be redelivered", map[string]any{
				"offset": msg.Offset,
			})
			return
		}
	}
}

func sleep(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}

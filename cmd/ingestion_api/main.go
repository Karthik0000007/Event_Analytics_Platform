// Command ingestion_api accepts events over HTTP and publishes them to Kafka.
// It also serves the read endpoints used by the dashboard.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Karthik0000007/Event_Analytics_Platform/internal/api"
	"github.com/Karthik0000007/Event_Analytics_Platform/internal/config"
	"github.com/Karthik0000007/Event_Analytics_Platform/internal/logging"
	"github.com/Karthik0000007/Event_Analytics_Platform/internal/messaging"
	"github.com/Karthik0000007/Event_Analytics_Platform/internal/storage"
)

const shutdownTimeout = 10 * time.Second

func main() {
	cfg := config.Load()
	logger := logging.New(cfg.ServiceName)

	if err := run(cfg, logger); err != nil {
		logger.Error("service stopped with error", map[string]any{"error": err.Error()})
		os.Exit(1)
	}
}

func run(cfg *config.Config, logger *logging.Logger) error {
	// ctx is cancelled on Ctrl+C (SIGINT) or `docker stop` (SIGTERM).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := storage.New(cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	producer := messaging.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopic)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           api.NewRouter(producer, db, logger),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		logger.Info("listening", map[string]any{"port": cfg.Port, "topic": cfg.KafkaTopic})
		serveErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		closeAll(logger, producer, db)
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
	}

	// Stop accepting new requests and let in-flight requests finish.
	logger.Info("shutting down", map[string]any{"timeout": shutdownTimeout.String()})
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	shutdownErr := server.Shutdown(shutdownCtx)

	// Close the producer first (flushes pending writes), then the database.
	closeAll(logger, producer, db)

	if shutdownErr != nil && !errors.Is(shutdownErr, http.ErrServerClosed) {
		return fmt.Errorf("graceful shutdown: %w", shutdownErr)
	}
	logger.Info("stopped cleanly", nil)
	return nil
}

func closeAll(logger *logging.Logger, producer *messaging.Producer, db *storage.DB) {
	if err := producer.Close(); err != nil {
		logger.Error("closing kafka producer", map[string]any{"error": err.Error()})
	}
	if err := db.Close(); err != nil {
		logger.Error("closing database", map[string]any{"error": err.Error()})
	}
}

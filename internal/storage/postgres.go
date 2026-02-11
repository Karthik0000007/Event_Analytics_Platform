package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

// New opens a connection pool to PostgreSQL and verifies connectivity.
func New(dsn string) (*DB, error) {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return &DB{conn: conn}, nil
}

// InsertEvent performs an idempotent upsert keyed on event_id (PRIMARY KEY).
// Duplicate replays are safely ignored via ON CONFLICT DO NOTHING.
func (db *DB) InsertEvent(ctx context.Context, eventID, eventType string, payload json.RawMessage) error {
	query := `
		INSERT INTO events (event_id, event_type, payload, received_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (event_id) DO NOTHING
	`
	_, err := db.conn.ExecContext(ctx, query, eventID, eventType, payload)
	if err != nil {
		return fmt.Errorf("insert event %s: %w", eventID, err)
	}
	return nil
}

// Close shuts down the connection pool.
func (db *DB) Close() error {
	return db.conn.Close()
}

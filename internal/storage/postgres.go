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

// GetEvents returns a paginated, filterable list of events.
func (db *DB) GetEvents(ctx context.Context, eventType string, from, to *time.Time, limit, offset int) ([]Event, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	idx := 1

	if eventType != "" {
		where += fmt.Sprintf(" AND event_type = $%d", idx)
		args = append(args, eventType)
		idx++
	}
	if from != nil {
		where += fmt.Sprintf(" AND received_at >= $%d", idx)
		args = append(args, *from)
		idx++
	}
	if to != nil {
		where += fmt.Sprintf(" AND received_at <= $%d", idx)
		args = append(args, *to)
		idx++
	}

	// Count total matching rows
	countQuery := "SELECT COUNT(*) FROM events " + where
	var total int
	if err := db.conn.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count events: %w", err)
	}

	// Fetch page
	query := fmt.Sprintf(
		"SELECT event_id, event_type, payload, received_at FROM events %s ORDER BY received_at DESC LIMIT $%d OFFSET $%d",
		where, idx, idx+1,
	)
	args = append(args, limit, offset)

	rows, err := db.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.EventID, &e.EventType, &e.Payload, &e.ReceivedAt); err != nil {
			return nil, 0, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, e)
	}
	return events, total, rows.Err()
}

// GetEvent returns a single event by ID.
func (db *DB) GetEvent(ctx context.Context, eventID string) (*Event, error) {
	query := `SELECT event_id, event_type, payload, received_at FROM events WHERE event_id = $1`
	var e Event
	err := db.conn.QueryRowContext(ctx, query, eventID).Scan(&e.EventID, &e.EventType, &e.Payload, &e.ReceivedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get event %s: %w", eventID, err)
	}
	return &e, nil
}

// Summary holds aggregate stats.
type Summary struct {
	TotalEvents int      `json:"total_events"`
	TodayEvents int      `json:"today_events"`
	EventTypes  int      `json:"event_types"`
	TopTypes    []string `json:"top_types"`
}

// GetSummary returns aggregate statistics.
func (db *DB) GetSummary(ctx context.Context) (*Summary, error) {
	s := &Summary{}

	if err := db.conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM events").Scan(&s.TotalEvents); err != nil {
		return nil, fmt.Errorf("count total: %w", err)
	}
	if err := db.conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM events WHERE received_at >= CURRENT_DATE").Scan(&s.TodayEvents); err != nil {
		return nil, fmt.Errorf("count today: %w", err)
	}
	if err := db.conn.QueryRowContext(ctx, "SELECT COUNT(DISTINCT event_type) FROM events").Scan(&s.EventTypes); err != nil {
		return nil, fmt.Errorf("count types: %w", err)
	}

	rows, err := db.conn.QueryContext(ctx, "SELECT event_type FROM events GROUP BY event_type ORDER BY COUNT(*) DESC LIMIT 5")
	if err != nil {
		return nil, fmt.Errorf("top types: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		s.TopTypes = append(s.TopTypes, t)
	}
	return s, rows.Err()
}

// TypeCount holds a type and its count.
type TypeCount struct {
	EventType string `json:"event_type"`
	Count     int    `json:"count"`
}

// GetTypeCounts returns event counts grouped by type.
func (db *DB) GetTypeCounts(ctx context.Context) ([]TypeCount, error) {
	rows, err := db.conn.QueryContext(ctx, "SELECT event_type, COUNT(*) FROM events GROUP BY event_type ORDER BY COUNT(*) DESC")
	if err != nil {
		return nil, fmt.Errorf("type counts: %w", err)
	}
	defer rows.Close()
	var counts []TypeCount
	for rows.Next() {
		var tc TypeCount
		if err := rows.Scan(&tc.EventType, &tc.Count); err != nil {
			return nil, err
		}
		counts = append(counts, tc)
	}
	return counts, rows.Err()
}

// TimelinePoint holds a time bucket and its event count.
type TimelinePoint struct {
	Bucket time.Time `json:"bucket"`
	Count  int       `json:"count"`
}

// GetTimeline returns event counts grouped by hour for the last 24 hours.
func (db *DB) GetTimeline(ctx context.Context, hours int) ([]TimelinePoint, error) {
	query := `
		SELECT date_trunc('hour', received_at) AS bucket, COUNT(*)
		FROM events
		WHERE received_at >= NOW() - ($1 || ' hours')::INTERVAL
		GROUP BY bucket
		ORDER BY bucket
	`
	rows, err := db.conn.QueryContext(ctx, query, fmt.Sprintf("%d", hours))
	if err != nil {
		return nil, fmt.Errorf("timeline: %w", err)
	}
	defer rows.Close()
	var points []TimelinePoint
	for rows.Next() {
		var tp TimelinePoint
		if err := rows.Scan(&tp.Bucket, &tp.Count); err != nil {
			return nil, err
		}
		points = append(points, tp)
	}
	return points, rows.Err()
}

// Event represents a stored event row.
type Event struct {
	EventID    string          `json:"event_id"`
	EventType  string          `json:"event_type"`
	Payload    json.RawMessage `json:"payload"`
	ReceivedAt time.Time       `json:"received_at"`
}

// Close shuts down the connection pool.
func (db *DB) Close() error {
	return db.conn.Close()
}

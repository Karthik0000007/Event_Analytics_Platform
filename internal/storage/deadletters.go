package storage

import (
	"context"
	"fmt"
	"time"
)

// DeadLetter is one failed message, as stored in the dead_letters table.
type DeadLetter struct {
	ID             int64     `json:"id"`
	OriginalKey    string    `json:"original_key"`
	OriginalValue  string    `json:"original_value"`
	ErrorKind      string    `json:"error_kind"`
	ErrorMessage   string    `json:"error_message"`
	RetryCount     int       `json:"retry_count"`
	FailedAt       time.Time `json:"failed_at"`
	KafkaPartition int       `json:"kafka_partition"`
	KafkaOffset    int64     `json:"kafka_offset"`
}

// Page size limits for GET /v1/dlq.
const (
	DefaultDeadLetterLimit = 50
	MaxDeadLetterLimit     = 100
)

// insertDeadLetterSQL is idempotent: a DLQ message replayed after a crash
// has the same partition and offset, so the second insert does nothing.
const insertDeadLetterSQL = `
	INSERT INTO dead_letters
		(original_key, original_value, error_kind, error_message, retry_count, failed_at, kafka_partition, kafka_offset)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (kafka_partition, kafka_offset) DO NOTHING`

// InsertDeadLetter stores one dead letter.
func (db *DB) InsertDeadLetter(ctx context.Context, d DeadLetter) error {
	_, err := db.conn.ExecContext(ctx, insertDeadLetterSQL, deadLetterArgs(d)...)
	if err != nil {
		return fmt.Errorf("insert dead letter at offset %d: %w", d.KafkaOffset, err)
	}
	return nil
}

// deadLetterArgs returns the insert parameters in column order.
func deadLetterArgs(d DeadLetter) []any {
	return []any{d.OriginalKey, d.OriginalValue, d.ErrorKind, d.ErrorMessage,
		d.RetryCount, d.FailedAt, d.KafkaPartition, d.KafkaOffset}
}

// ClampPage turns user-supplied paging values into safe ones.
func ClampPage(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = DefaultDeadLetterLimit
	}
	if limit > MaxDeadLetterLimit {
		limit = MaxDeadLetterLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// listDeadLettersQuery builds the page query, newest first.
// Paging values are passed as parameters, never formatted into the SQL.
func listDeadLettersQuery(limit, offset int) (string, []any) {
	limit, offset = ClampPage(limit, offset)
	query := `
		SELECT id, original_key, original_value, error_kind, error_message,
		       retry_count, failed_at, kafka_partition, kafka_offset
		FROM dead_letters
		ORDER BY failed_at DESC, id DESC
		LIMIT $1 OFFSET $2`
	return query, []any{limit, offset}
}

// ListDeadLetters returns one page of dead letters and the total count.
func (db *DB) ListDeadLetters(ctx context.Context, limit, offset int) ([]DeadLetter, int, error) {
	var total int
	if err := db.conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM dead_letters").Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count dead letters: %w", err)
	}

	query, args := listDeadLettersQuery(limit, offset)
	rows, err := db.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list dead letters: %w", err)
	}
	defer rows.Close()

	letters := []DeadLetter{}
	for rows.Next() {
		var d DeadLetter
		if err := rows.Scan(&d.ID, &d.OriginalKey, &d.OriginalValue, &d.ErrorKind, &d.ErrorMessage,
			&d.RetryCount, &d.FailedAt, &d.KafkaPartition, &d.KafkaOffset); err != nil {
			return nil, 0, fmt.Errorf("scan dead letter: %w", err)
		}
		letters = append(letters, d)
	}
	return letters, total, rows.Err()
}

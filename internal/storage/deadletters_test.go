package storage

import (
	"strings"
	"testing"
	"time"
)

func TestClampPage(t *testing.T) {
	tests := []struct {
		name                  string
		limit, offset         int
		wantLimit, wantOffset int
	}{
		{"defaults when empty", 0, 0, DefaultDeadLetterLimit, 0},
		{"keeps valid values", 20, 40, 20, 40},
		{"caps large limit", 5000, 0, MaxDeadLetterLimit, 0},
		{"negative limit uses default", -1, 0, DefaultDeadLetterLimit, 0},
		{"negative offset becomes 0", 10, -5, 10, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l, o := ClampPage(tc.limit, tc.offset)
			if l != tc.wantLimit || o != tc.wantOffset {
				t.Errorf("ClampPage(%d, %d) = (%d, %d), want (%d, %d)", tc.limit, tc.offset, l, o, tc.wantLimit, tc.wantOffset)
			}
		})
	}
}

func TestListDeadLettersQuery_UsesParameters(t *testing.T) {
	query, args := listDeadLettersQuery(5000, -3)

	if !strings.Contains(query, "LIMIT $1 OFFSET $2") {
		t.Errorf("paging must use placeholders, got query:\n%s", query)
	}
	if !strings.Contains(query, "ORDER BY failed_at DESC") {
		t.Error("dead letters must be listed newest first")
	}
	if len(args) != 2 || args[0] != MaxDeadLetterLimit || args[1] != 0 {
		t.Errorf("expected clamped args [%d 0], got %v", MaxDeadLetterLimit, args)
	}
}

func TestInsertDeadLetterSQL_IsIdempotent(t *testing.T) {
	if !strings.Contains(insertDeadLetterSQL, "ON CONFLICT (kafka_partition, kafka_offset) DO NOTHING") {
		t.Error("a replayed DLQ message must not create a second row")
	}
	if got := strings.Count(insertDeadLetterSQL, "$"); got != 8 {
		t.Errorf("expected 8 placeholders, got %d", got)
	}
	args := deadLetterArgs(DeadLetter{OriginalKey: "k", FailedAt: time.Now(), KafkaOffset: 9})
	if len(args) != 8 || args[0] != "k" || args[7] != int64(9) {
		t.Errorf("args do not match the column order: %v", args)
	}
}

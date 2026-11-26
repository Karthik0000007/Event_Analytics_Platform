package logging

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestLogger_WritesOneJSONObjectPerLine(t *testing.T) {
	var buf bytes.Buffer
	l := NewWriter("test-service", &buf)

	fields := map[string]any{"event_id": "abc"}
	l.Info("hello", fields)

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("log line is not valid JSON: %v (%q)", err, buf.String())
	}
	if got["level"] != "INFO" || got["service"] != "test-service" || got["message"] != "hello" {
		t.Errorf("unexpected log entry: %v", got)
	}
	if got["event_id"] != "abc" {
		t.Errorf("custom field missing: %v", got)
	}
	if _, ok := fields["level"]; ok {
		t.Error("logger must not modify the caller's map")
	}
}

func TestLogger_AcceptsNilFields(t *testing.T) {
	var buf bytes.Buffer
	NewWriter("svc", &buf).Error("boom", nil)
	if !json.Valid(bytes.TrimSpace(buf.Bytes())) {
		t.Errorf("expected valid JSON, got %q", buf.String())
	}
}

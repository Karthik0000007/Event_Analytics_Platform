package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Karthik0000007/Event_Analytics_Platform/internal/logging"
)

// fakePublisher records what the handler publishes, so no Kafka is needed.
type fakePublisher struct {
	err       error
	published []EventRequest
}

func (f *fakePublisher) Publish(_ context.Context, _ string, event any) error {
	if f.err != nil {
		return f.err
	}
	f.published = append(f.published, event.(EventRequest))
	return nil
}

const validID = "550e8400-e29b-41d4-a716-446655440000"

func postEvent(t *testing.T, pub Publisher, body string) *httptest.ResponseRecorder {
	t.Helper()
	h := HandleEvent(pub, logging.NewWriter("test", io.Discard))
	req := httptest.NewRequest(http.MethodPost, "/v1/events", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h(rec, req)
	return rec
}

func TestHandleEvent_Validation(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
	}{
		{"valid event", `{"event_id":"` + validID + `","event_type":"click","payload":{"page":"/"}}`, http.StatusAccepted, ""},
		{"valid without payload", `{"event_id":"` + validID + `","event_type":"click"}`, http.StatusAccepted, ""},
		{"invalid JSON", `{"event_id":`, http.StatusBadRequest, "valid JSON"},
		{"not an object", `[1,2,3]`, http.StatusBadRequest, "valid JSON"},
		{"two JSON objects", `{"event_id":"` + validID + `","event_type":"a"}{}`, http.StatusBadRequest, "single JSON object"},
		{"missing event_id", `{"event_type":"click"}`, http.StatusBadRequest, "event_id is required"},
		{"event_id not a UUID", `{"event_id":"abc","event_type":"click"}`, http.StatusBadRequest, "UUID"},
		{"missing event_type", `{"event_id":"` + validID + `"}`, http.StatusBadRequest, "event_type is required"},
		{"blank event_type", `{"event_id":"` + validID + `","event_type":"   "}`, http.StatusBadRequest, "event_type is required"},
		{"event_type 64 chars", `{"event_id":"` + validID + `","event_type":"` + strings.Repeat("a", 64) + `"}`, http.StatusAccepted, ""},
		{"event_type 65 chars", `{"event_id":"` + validID + `","event_type":"` + strings.Repeat("a", 65) + `"}`, http.StatusBadRequest, "at most 64"},
		{"body over 64 KB", `{"event_id":"` + validID + `","event_type":"big","payload":"` + strings.Repeat("x", MaxBodyBytes) + `"}`, http.StatusRequestEntityTooLarge, "64 KB"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pub := &fakePublisher{}
			rec := postEvent(t, pub, tc.body)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status: want %d, got %d (body %s)", tc.wantStatus, rec.Code, rec.Body.String())
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type: want application/json, got %q", ct)
			}

			var resp map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("response is not JSON: %v", err)
			}

			if tc.wantError == "" {
				if resp["event_id"] != validID || len(pub.published) != 1 {
					t.Errorf("expected event %s published once, got resp=%v published=%d", validID, resp, len(pub.published))
				}
				return
			}
			if !strings.Contains(resp["error"], tc.wantError) {
				t.Errorf("error: want it to contain %q, got %q", tc.wantError, resp["error"])
			}
			if len(pub.published) != 0 {
				t.Error("invalid input must not be published")
			}
		})
	}
}

func TestHandleEvent_MissingPayloadBecomesEmptyObject(t *testing.T) {
	pub := &fakePublisher{}
	postEvent(t, pub, `{"event_id":"`+validID+`","event_type":"click","payload":null}`)

	if len(pub.published) != 1 || string(pub.published[0].Payload) != "{}" {
		t.Fatalf("expected payload {} to be published, got %+v", pub.published)
	}
}

func TestHandleEvent_KafkaDownReturns503(t *testing.T) {
	pub := &fakePublisher{err: errors.New("kafka: broker not available")}
	rec := postEvent(t, pub, `{"event_id":"`+validID+`","event_type":"click"}`)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when publish fails, got %d", rec.Code)
	}
}

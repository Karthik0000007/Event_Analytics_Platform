package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Karthik0000007/Event_Analytics_Platform/internal/logging"
	"github.com/google/uuid"
)

const (
	// MaxBodyBytes is the largest request body POST /v1/events accepts.
	MaxBodyBytes = 64 << 10 // 64 KB
	// MaxEventTypeLen is the longest event_type accepted, in characters.
	MaxEventTypeLen = 64

	publishTimeout = 5 * time.Second
)

// EventRequest is the body of POST /v1/events. It is also the Kafka message value.
type EventRequest struct {
	EventID string          `json:"event_id"`
	Type    string          `json:"event_type"`
	Payload json.RawMessage `json:"payload"`
}

// Publisher sends an event to the message broker.
// *messaging.Producer implements it; tests use a fake.
type Publisher interface {
	Publish(ctx context.Context, key string, event any) error
}

// HandleEvent validates the request and publishes it to Kafka before answering.
// 202 means Kafka has stored the event, so a crash after the response cannot lose it.
func HandleEvent(publisher Publisher, logger *logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, status, msg := decodeEvent(w, r)
		if status != 0 {
			writeError(w, status, msg)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), publishTimeout)
		defer cancel()
		if err := publisher.Publish(ctx, req.EventID, req); err != nil {
			logger.Error("kafka publish failed", map[string]any{"event_id": req.EventID, "error": err.Error()})
			writeError(w, http.StatusServiceUnavailable, "event could not be queued, please retry")
			return
		}

		writeJSON(w, http.StatusAccepted, map[string]string{
			"status":   "accepted",
			"event_id": req.EventID,
		})
	}
}

// decodeEvent reads and validates the body. A non-zero status means the
// request is invalid and msg explains why.
func decodeEvent(w http.ResponseWriter, r *http.Request) (EventRequest, int, string) {
	var req EventRequest
	body := http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	dec := json.NewDecoder(body)

	if err := dec.Decode(&req); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			return req, http.StatusRequestEntityTooLarge, "request body must be at most 64 KB"
		}
		return req, http.StatusBadRequest, "request body must be valid JSON"
	}
	if dec.Decode(&struct{}{}) != io.EOF {
		return req, http.StatusBadRequest, "request body must contain a single JSON object"
	}

	if status, msg := validateEvent(req); status != 0 {
		return req, status, msg
	}
	if len(req.Payload) == 0 || bytes.Equal(req.Payload, []byte("null")) {
		req.Payload = json.RawMessage("{}")
	}
	return req, 0, ""
}

// validateEvent checks the required fields.
func validateEvent(req EventRequest) (int, string) {
	switch {
	case strings.TrimSpace(req.EventID) == "":
		return http.StatusBadRequest, "event_id is required"
	case uuid.Validate(req.EventID) != nil:
		return http.StatusBadRequest, "event_id must be a UUID"
	case strings.TrimSpace(req.Type) == "":
		return http.StatusBadRequest, "event_type is required"
	case utf8.RuneCountInString(req.Type) > MaxEventTypeLen:
		return http.StatusBadRequest, "event_type must be at most 64 characters"
	}
	return 0, ""
}

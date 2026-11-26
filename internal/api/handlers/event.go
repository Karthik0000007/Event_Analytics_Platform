package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Karthik0000007/Event_Analytics_Platform/internal/logging"
	"github.com/google/uuid"
)

type EventRequest struct {
	EventID string          `json:"event_id"`
	Type    string          `json:"event_type"`
	Payload json.RawMessage `json:"payload"`
}

// Publisher sends an event to the message broker.
// *messaging.Producer implements it; tests can use a fake.
type Publisher interface {
	Publish(ctx context.Context, key string, event any) error
}

func HandleEvent(publisher Publisher, logger *logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req EventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		if _, err := uuid.Parse(req.EventID); err != nil {
			http.Error(w, "invalid event_id", http.StatusBadRequest)
			return
		}

		// Publish event asynchronously to Kafka without blocking the response
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := publisher.Publish(ctx, req.EventID, req); err != nil {
				logger.Error("kafka publish failed", map[string]any{"event_id": req.EventID, "error": err.Error()})
			}
		}()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "accepted",
			"event_id": req.EventID,
			"message":  "Event accepted for processing",
		})
	}
}

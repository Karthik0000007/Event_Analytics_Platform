package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type EventRequest struct {
	EventID string          `json:"event_id"`
	Type    string          `json:"event_type"`
	Payload json.RawMessage `json:"payload"`
}

func HandleEvent(w http.ResponseWriter, r *http.Request) {
	var req EventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if _, err := uuid.Parse(req.EventID); err != nil {
		http.Error(w, "invalid event_id", http.StatusBadRequest)
		return
	}

	// Kafka publish will go here (Day-10)
	w.WriteHeader(http.StatusAccepted)
}

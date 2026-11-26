package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Karthik0000007/Event_Analytics_Platform/internal/storage"
)

// DeadLetterLister reads stored dead letters. *storage.DB implements it.
type DeadLetterLister interface {
	ListDeadLetters(ctx context.Context, limit, offset int) ([]storage.DeadLetter, int, error)
}

// DLQHandler serves GET /v1/dlq.
type DLQHandler struct {
	Store DeadLetterLister
}

// List handles GET /v1/dlq?limit=50&offset=0 (newest first).
func (h *DLQHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, offset = storage.ClampPage(limit, offset)

	letters, total, err := h.Store.ListDeadLetters(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list dead letters")
		return
	}
	if letters == nil {
		letters = []storage.DeadLetter{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"dead_letters": letters,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
	})
}

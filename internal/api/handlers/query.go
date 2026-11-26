package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Karthik0000007/Event_Analytics_Platform/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// QueryHandlers holds read-only handlers backed by the database.
type QueryHandlers struct {
	DB *storage.DB
}

// ListEvents handles GET /v1/events with optional query params:
//
//	?type=click&limit=50&offset=0&from=2026-01-01T00:00:00Z&to=2026-02-01T00:00:00Z
func (q *QueryHandlers) ListEvents(w http.ResponseWriter, r *http.Request) {
	eventType := r.URL.Query().Get("type")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var from, to *time.Time
	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			from = &t
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			to = &t
		}
	}

	events, total, err := q.DB.GetEvents(r.Context(), eventType, from, to, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query events")
		return
	}
	if events == nil {
		events = []storage.Event{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"events": events,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetEvent handles GET /v1/events/{id}
func (q *QueryHandlers) GetEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if uuid.Validate(id) != nil {
		writeError(w, http.StatusBadRequest, "event id must be a UUID")
		return
	}
	event, err := q.DB.GetEvent(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get event")
		return
	}
	if event == nil {
		writeError(w, http.StatusNotFound, "event not found")
		return
	}

	writeJSON(w, http.StatusOK, event)
}

// GetSummary handles GET /v1/analytics/summary
func (q *QueryHandlers) GetSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := q.DB.GetSummary(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get summary")
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

// GetTypeCounts handles GET /v1/analytics/types
func (q *QueryHandlers) GetTypeCounts(w http.ResponseWriter, r *http.Request) {
	counts, err := q.DB.GetTypeCounts(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get type counts")
		return
	}
	if counts == nil {
		counts = []storage.TypeCount{}
	}

	writeJSON(w, http.StatusOK, counts)
}

// GetTimeline handles GET /v1/analytics/timeline?hours=24
func (q *QueryHandlers) GetTimeline(w http.ResponseWriter, r *http.Request) {
	hours, _ := strconv.Atoi(r.URL.Query().Get("hours"))
	if hours <= 0 || hours > 720 {
		hours = 24
	}
	points, err := q.DB.GetTimeline(r.Context(), hours)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get timeline")
		return
	}
	if points == nil {
		points = []storage.TimelinePoint{}
	}

	writeJSON(w, http.StatusOK, points)
}

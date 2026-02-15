package api

import (
	"net/http"

	"github.com/Karthik0000007/Event_Analytics_Platform/internal/api/handlers"
	"github.com/Karthik0000007/Event_Analytics_Platform/internal/api/middleware"
	"github.com/Karthik0000007/Event_Analytics_Platform/internal/health"
	"github.com/Karthik0000007/Event_Analytics_Platform/internal/messaging"
	"github.com/Karthik0000007/Event_Analytics_Platform/internal/storage"
	"github.com/go-chi/chi/v5"
)

func NewRouter(producer *messaging.Producer, db *storage.DB) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logging)
	r.Use(corsMiddleware)

	r.Get("/healthz", health.Liveness)
	r.Get("/readyz", health.Readiness)

	qh := &handlers.QueryHandlers{DB: db}

	r.Route("/v1", func(r chi.Router) {
		// Write
		r.Post("/events", handlers.HandleEvent(producer))

		// Read
		r.Get("/events", qh.ListEvents)
		r.Get("/events/{id}", qh.GetEvent)

		// Analytics
		r.Get("/analytics/summary", qh.GetSummary)
		r.Get("/analytics/types", qh.GetTypeCounts)
		r.Get("/analytics/timeline", qh.GetTimeline)
	})

	return r
}

// corsMiddleware allows the frontend dev server to reach the API.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-ID")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

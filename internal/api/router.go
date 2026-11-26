package api

import (
	"net/http"

	"github.com/Karthik0000007/Event_Analytics_Platform/internal/api/handlers"
	"github.com/Karthik0000007/Event_Analytics_Platform/internal/api/middleware"
	"github.com/Karthik0000007/Event_Analytics_Platform/internal/health"
	"github.com/Karthik0000007/Event_Analytics_Platform/internal/logging"
	"github.com/Karthik0000007/Event_Analytics_Platform/internal/storage"
	"github.com/go-chi/chi/v5"
)

// NewRouter wires every route to its handler.
// publisher sends accepted events to Kafka; db serves the read endpoints.
func NewRouter(publisher handlers.Publisher, db *storage.DB, logger *logging.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logging(logger))
	r.Use(corsMiddleware)

	r.Get("/healthz", health.Liveness)
	r.Get("/readyz", health.Readiness(db))

	qh := &handlers.QueryHandlers{DB: db}

	r.Route("/v1", func(r chi.Router) {
		// Write
		r.Post("/events", handlers.HandleEvent(publisher, logger))

		// Read
		r.Get("/events", qh.ListEvents)
		r.Get("/events/{id}", qh.GetEvent)

		// Analytics
		r.Get("/analytics/summary", qh.GetSummary)
		r.Get("/analytics/types", qh.GetTypeCounts)
		r.Get("/analytics/timeline", qh.GetTimeline)

		// Dead letters, written to Postgres by cmd/dlq-recorder
		r.Get("/dlq", (&handlers.DLQHandler{Store: db}).List)
	})

	return r
}

// corsMiddleware lets the Vite dev server (port 3000) call the API directly.
// In Docker the dashboard goes through the nginx proxy, so CORS is not needed there.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-ID")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

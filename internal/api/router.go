package api

import (
	"net/http"

	"github.com/Karthik0000007/Event_Analytics_Platform/internal/api/handlers"
	"github.com/Karthik0000007/Event_Analytics_Platform/internal/health"
	"github.com/go-chi/chi/v5"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/healthz", health.Liveness)
	r.Get("/readyz", health.Readiness)

	r.Route("/v1", func(r chi.Router) {
		r.Post("/events", handlers.HandleEvent)
	})

	return r
}

package middleware

import (
	"net/http"
	"time"

	"github.com/Karthik0000007/Event_Analytics_Platform/internal/logging"
)

func Logging(logger *logging.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			reqID, _ := r.Context().Value(RequestIDKey).(string)

			logger.Info("request started", map[string]any{
				"request_id": reqID,
				"method":     r.Method,
				"path":       r.URL.Path,
			})

			next.ServeHTTP(w, r)

			logger.Info("request completed", map[string]any{
				"request_id": reqID,
				"duration_ms": time.Since(start).Milliseconds(),
			})
		})
	}
}

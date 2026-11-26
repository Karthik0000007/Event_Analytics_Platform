package middleware

import (
	"net/http"
	"time"

	"github.com/Karthik0000007/Event_Analytics_Platform/internal/logging"
)

// statusRecorder remembers the status code a handler wrote.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// Logging writes one log line per request with method, path, status and duration.
func Logging(logger *logging.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)

			reqID, _ := r.Context().Value(RequestIDKey).(string)
			logger.Info("request completed", map[string]any{
				"request_id":  reqID,
				"method":      r.Method,
				"path":        r.URL.Path,
				"status":      rec.status,
				"duration_ms": time.Since(start).Milliseconds(),
			})
		})
	}
}

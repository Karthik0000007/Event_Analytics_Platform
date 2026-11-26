// Package health serves the liveness and readiness probes.
//
// Liveness (/healthz) answers "is the process alive?" and never checks
// dependencies, so a database outage does not make Kubernetes restart the API.
// Readiness (/readyz) answers "can this instance serve all its routes?" and
// pings PostgreSQL, because the read endpoints need it.
package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Pinger is anything that can check its own connection, such as *storage.DB.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Liveness always returns 200 while the process can serve HTTP.
func Liveness(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Readiness returns 200 when the database answers a ping, otherwise 503.
func Readiness(db Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status":   "not ready",
				"database": "unreachable",
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready", "database": "ok"})
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

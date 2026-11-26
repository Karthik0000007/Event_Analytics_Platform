package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakePinger struct{ err error }

func (f fakePinger) Ping(context.Context) error { return f.err }

func TestLiveness_AlwaysOK(t *testing.T) {
	rec := httptest.NewRecorder()
	Liveness(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestReadiness(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"database up", nil, http.StatusOK},
		{"database down", errors.New("connection refused"), http.StatusServiceUnavailable},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			Readiness(fakePinger{tc.err})(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
			if rec.Code != tc.want {
				t.Errorf("expected %d, got %d", tc.want, rec.Code)
			}
		})
	}
}

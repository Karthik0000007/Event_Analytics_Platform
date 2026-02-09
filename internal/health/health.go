package health

import "net/http"

func Liveness(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func Readiness(w http.ResponseWriter, _ *http.Request) {
	// Kafka / DB checks will come later
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ready"))
}

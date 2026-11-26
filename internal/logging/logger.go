// Package logging writes one JSON object per line, so logs are easy to
// read in `docker compose logs` and easy to parse by log tools.
package logging

import (
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"
)

// Logger writes structured JSON log lines.
type Logger struct {
	mu      sync.Mutex
	out     io.Writer
	Service string
}

// New returns a logger that writes to stdout.
func New(service string) *Logger {
	return NewWriter(service, os.Stdout)
}

// NewWriter returns a logger that writes to w (useful in tests).
func NewWriter(service string, w io.Writer) *Logger {
	return &Logger{out: w, Service: service}
}

// Info logs a normal event.
func (l *Logger) Info(msg string, fields map[string]any) {
	l.log("INFO", msg, fields)
}

// Error logs a failure.
func (l *Logger) Error(msg string, fields map[string]any) {
	l.log("ERROR", msg, fields)
}

func (l *Logger) log(level, msg string, fields map[string]any) {
	entry := make(map[string]any, len(fields)+4)
	for k, v := range fields {
		entry[k] = v
	}
	entry["time"] = time.Now().UTC().Format(time.RFC3339Nano)
	entry["level"] = level
	entry["service"] = l.Service
	entry["message"] = msg

	line, err := json.Marshal(entry)
	if err != nil {
		line = []byte(`{"level":"ERROR","message":"could not encode log entry"}`)
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = l.out.Write(append(line, '\n'))
}

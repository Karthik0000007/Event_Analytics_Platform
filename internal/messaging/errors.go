package messaging

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"syscall"
)

// ──────────────────────────────────────────────────────────────
// Error classification: transient vs permanent
// ──────────────────────────────────────────────────────────────

// ErrorKind categorises a processing failure so the consumer can
// decide whether to retry (transient) or route to DLQ (permanent).
type ErrorKind int

const (
	ErrTransient ErrorKind = iota // e.g. DB timeout, Kafka broker hiccup
	ErrPermanent                  // e.g. malformed JSON, constraint violation
)

func (k ErrorKind) String() string {
	switch k {
	case ErrTransient:
		return "transient"
	case ErrPermanent:
		return "permanent"
	default:
		return "unknown"
	}
}

// ProcessingError wraps an underlying error with classification metadata.
type ProcessingError struct {
	Kind    ErrorKind
	Cause   error
	Message string
}

func (e *ProcessingError) Error() string {
	return fmt.Sprintf("[%s] %s: %v", e.Kind, e.Message, e.Cause)
}

func (e *ProcessingError) Unwrap() error { return e.Cause }

// NewTransient creates a retryable error.
func NewTransient(msg string, cause error) *ProcessingError {
	return &ProcessingError{Kind: ErrTransient, Cause: cause, Message: msg}
}

// NewPermanent creates a non-retryable error.
func NewPermanent(msg string, cause error) *ProcessingError {
	return &ProcessingError{Kind: ErrPermanent, Cause: cause, Message: msg}
}

// Classify inspects an error and returns whether it is transient or permanent.
// Network errors, timeouts, connection resets → transient.
// Everything else (bad JSON, constraint violations) → permanent.
func Classify(err error) ErrorKind {
	if err == nil {
		return ErrTransient // should not be called with nil, but safe default
	}

	// Explicit classification already present
	var pe *ProcessingError
	if errors.As(err, &pe) {
		return pe.Kind
	}

	// Network / IO errors are transient
	var netErr net.Error
	if errors.As(err, &netErr) {
		return ErrTransient
	}
	if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
		return ErrTransient
	}
	if errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ECONNRESET) {
		return ErrTransient
	}

	// Postgres unique-violation / check-constraint → permanent
	msg := err.Error()
	if strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "violates check constraint") ||
		strings.Contains(msg, "invalid input syntax") {
		return ErrPermanent
	}

	// Connection-related postgres errors → transient
	if strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "too many clients") {
		return ErrTransient
	}

	// Unknown errors default to transient so we retry rather than discard.
	return ErrTransient
}

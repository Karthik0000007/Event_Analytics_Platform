package messaging

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"syscall"
)

// ErrorKind says whether a failure is worth retrying.
type ErrorKind int

const (
	ErrTransient ErrorKind = iota // e.g. database down, timeout, broker hiccup
	ErrPermanent                  // e.g. malformed JSON, a value Postgres rejects
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

// ProcessingError wraps an error with an explicit classification.
type ProcessingError struct {
	Kind    ErrorKind
	Cause   error
	Message string
}

func (e *ProcessingError) Error() string {
	if e.Cause == nil {
		return fmt.Sprintf("[%s] %s", e.Kind, e.Message)
	}
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

// sqlStateError is implemented by *pq.Error. Using an interface keeps this
// package free of a direct dependency on the Postgres driver.
type sqlStateError interface {
	SQLState() string
}

// Classify decides whether err is transient (retry) or permanent (DLQ).
// Order of checks:
//  1. an explicit ProcessingError
//  2. the Postgres SQLSTATE code, when the error has one
//  3. network and I/O errors
//  4. known message text, for errors without a code
//  5. anything unknown is transient, because retrying is safer than discarding
func Classify(err error) ErrorKind {
	if err == nil {
		return ErrTransient
	}

	var pe *ProcessingError
	if errors.As(err, &pe) {
		return pe.Kind
	}

	var se sqlStateError
	if errors.As(err, &se) {
		return classifySQLState(se.SQLState())
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return ErrTransient
	}
	if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) ||
		errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ECONNRESET) {
		return ErrTransient
	}

	msg := err.Error()
	if strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "violates check constraint") ||
		strings.Contains(msg, "invalid input syntax") {
		return ErrPermanent
	}
	return ErrTransient
}

// classifySQLState maps a Postgres error code to a kind. The first two
// characters are the error class:
//
//	22 data exception (bad value, e.g. \u0000 inside JSONB) -> permanent
//	23 integrity constraint violation                        -> permanent
//	anything else (08 connection, 53 resources,
//	57 admin shutdown, 40 deadlock, ...)                     -> transient
func classifySQLState(code string) ErrorKind {
	if len(code) < 2 {
		return ErrTransient
	}
	switch code[:2] {
	case "22", "23":
		return ErrPermanent
	default:
		return ErrTransient
	}
}

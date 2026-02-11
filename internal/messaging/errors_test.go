package messaging

import (
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
	"testing"
)

// ──────────────────────────────────────────────────────────────
// Error classification tests
// ──────────────────────────────────────────────────────────────

func TestClassify_ExplicitProcessingError(t *testing.T) {
	perm := NewPermanent("bad input", errors.New("fail"))
	if Classify(perm) != ErrPermanent {
		t.Error("expected permanent for explicit permanent error")
	}
	trans := NewTransient("timeout", errors.New("fail"))
	if Classify(trans) != ErrTransient {
		t.Error("expected transient for explicit transient error")
	}
}

func TestClassify_WrappedProcessingError(t *testing.T) {
	inner := NewPermanent("bad", errors.New("x"))
	wrapped := fmt.Errorf("outer: %w", inner)
	if Classify(wrapped) != ErrPermanent {
		t.Error("expected permanent for wrapped permanent error")
	}
}

func TestClassify_NetworkErrors(t *testing.T) {
	// net.Error (timeout)
	netErr := &net.OpError{Op: "read", Err: errors.New("timeout")}
	if Classify(netErr) != ErrTransient {
		t.Error("expected transient for net.OpError")
	}

	// EOF
	if Classify(io.EOF) != ErrTransient {
		t.Error("expected transient for io.EOF")
	}
	if Classify(io.ErrUnexpectedEOF) != ErrTransient {
		t.Error("expected transient for io.ErrUnexpectedEOF")
	}

	// Connection refused
	if Classify(syscall.ECONNREFUSED) != ErrTransient {
		t.Error("expected transient for ECONNREFUSED")
	}
	if Classify(syscall.ECONNRESET) != ErrTransient {
		t.Error("expected transient for ECONNRESET")
	}
}

func TestClassify_PostgresConstraintViolation(t *testing.T) {
	err := errors.New(`pq: duplicate key value violates unique constraint "events_pkey"`)
	if Classify(err) != ErrPermanent {
		t.Error("expected permanent for unique constraint violation")
	}
}

func TestClassify_PostgresInvalidSyntax(t *testing.T) {
	err := errors.New(`pq: invalid input syntax for type uuid: "not-a-uuid"`)
	if Classify(err) != ErrPermanent {
		t.Error("expected permanent for invalid input syntax")
	}
}

func TestClassify_ConnectionTimeout(t *testing.T) {
	err := errors.New("dial tcp 127.0.0.1:5432: connection refused")
	if Classify(err) != ErrTransient {
		t.Error("expected transient for connection refused string")
	}

	err2 := errors.New("context deadline exceeded (timeout)")
	if Classify(err2) != ErrTransient {
		t.Error("expected transient for timeout string")
	}
}

func TestClassify_UnknownDefaultsToTransient(t *testing.T) {
	err := errors.New("some obscure error we haven't seen before")
	if Classify(err) != ErrTransient {
		t.Error("expected unknown errors to default to transient")
	}
}

func TestClassify_NilReturnsTransient(t *testing.T) {
	if Classify(nil) != ErrTransient {
		t.Error("expected nil error to return transient")
	}
}

func TestProcessingError_Unwrap(t *testing.T) {
	cause := errors.New("root cause")
	pe := NewPermanent("wrapper", cause)
	if !errors.Is(pe, cause) {
		t.Error("expected Unwrap to expose root cause")
	}
}

func TestErrorKind_String(t *testing.T) {
	if ErrTransient.String() != "transient" {
		t.Error("unexpected string for ErrTransient")
	}
	if ErrPermanent.String() != "permanent" {
		t.Error("unexpected string for ErrPermanent")
	}
	if ErrorKind(99).String() != "unknown" {
		t.Error("unexpected string for unknown ErrorKind")
	}
}

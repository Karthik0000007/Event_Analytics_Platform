package messaging

import (
	"context"
	"math"
	"math/rand"
	"time"
)

// ──────────────────────────────────────────────────────────────
// Bounded retry with exponential back-off + jitter
// ──────────────────────────────────────────────────────────────

// RetryConfig controls the retry discipline for transient failures.
type RetryConfig struct {
	MaxRetries  int           // Hard ceiling – after this, route to DLQ.
	BaseDelay   time.Duration // Initial back-off delay.
	MaxDelay    time.Duration // Cap so we don't sleep forever.
	Multiplier  float64       // Exponential factor (typically 2.0).
	JitterRatio float64       // 0.0–1.0; fraction of delay randomised.
}

// DefaultRetryConfig returns sane defaults for event processing.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:  5,
		BaseDelay:   200 * time.Millisecond,
		MaxDelay:    10 * time.Second,
		Multiplier:  2.0,
		JitterRatio: 0.3,
	}
}

// BackoffDelay calculates the delay for attempt n (0-indexed).
func (rc RetryConfig) BackoffDelay(attempt int) time.Duration {
	delay := float64(rc.BaseDelay) * math.Pow(rc.Multiplier, float64(attempt))
	if delay > float64(rc.MaxDelay) {
		delay = float64(rc.MaxDelay)
	}

	// Add jitter: ±JitterRatio of the computed delay
	jitter := delay * rc.JitterRatio * (rand.Float64()*2 - 1) // [-ratio, +ratio]
	delay += jitter
	if delay < 0 {
		delay = float64(rc.BaseDelay)
	}

	return time.Duration(delay)
}

// ShouldRetry returns true when the error is transient and we haven't
// exhausted our retry budget.
func (rc RetryConfig) ShouldRetry(kind ErrorKind, attempt int) bool {
	return kind == ErrTransient && attempt < rc.MaxRetries
}

// Sleep blocks for the calculated back-off duration, returning early
// if the context is cancelled.
func (rc RetryConfig) Sleep(ctx context.Context, attempt int) error {
	d := rc.BackoffDelay(attempt)
	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

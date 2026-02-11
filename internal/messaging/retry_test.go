package messaging

import (
	"context"
	"testing"
	"time"
)

// ──────────────────────────────────────────────────────────────
// Retry discipline tests
// ──────────────────────────────────────────────────────────────

func TestRetryConfig_ShouldRetry_TransientWithinBudget(t *testing.T) {
	rc := DefaultRetryConfig()
	for i := 0; i < rc.MaxRetries; i++ {
		if !rc.ShouldRetry(ErrTransient, i) {
			t.Errorf("expected ShouldRetry=true for attempt %d", i)
		}
	}
}

func TestRetryConfig_ShouldRetry_ExhaustedBudget(t *testing.T) {
	rc := DefaultRetryConfig()
	if rc.ShouldRetry(ErrTransient, rc.MaxRetries) {
		t.Error("expected ShouldRetry=false when budget exhausted")
	}
	if rc.ShouldRetry(ErrTransient, rc.MaxRetries+1) {
		t.Error("expected ShouldRetry=false beyond budget")
	}
}

func TestRetryConfig_ShouldRetry_PermanentNeverRetries(t *testing.T) {
	rc := DefaultRetryConfig()
	if rc.ShouldRetry(ErrPermanent, 0) {
		t.Error("expected ShouldRetry=false for permanent error even at attempt 0")
	}
}

func TestRetryConfig_BackoffDelay_Increases(t *testing.T) {
	rc := RetryConfig{
		MaxRetries:  5,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    10 * time.Second,
		Multiplier:  2.0,
		JitterRatio: 0, // disable jitter for deterministic testing
	}

	prev := time.Duration(0)
	for i := 0; i < 5; i++ {
		d := rc.BackoffDelay(i)
		if d <= prev && i > 0 {
			t.Errorf("expected delay to increase: attempt %d got %v, prev %v", i, d, prev)
		}
		prev = d
	}
}

func TestRetryConfig_BackoffDelay_CappedAtMax(t *testing.T) {
	rc := RetryConfig{
		MaxRetries:  10,
		BaseDelay:   1 * time.Second,
		MaxDelay:    5 * time.Second,
		Multiplier:  3.0,
		JitterRatio: 0,
	}

	d := rc.BackoffDelay(10)
	if d > rc.MaxDelay {
		t.Errorf("expected delay capped at %v, got %v", rc.MaxDelay, d)
	}
}

func TestRetryConfig_BackoffDelay_WithJitter(t *testing.T) {
	rc := RetryConfig{
		MaxRetries:  5,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    10 * time.Second,
		Multiplier:  2.0,
		JitterRatio: 0.5,
	}

	// Run multiple times; values should vary
	seen := make(map[time.Duration]bool)
	for i := 0; i < 20; i++ {
		d := rc.BackoffDelay(2)
		seen[d] = true
	}

	if len(seen) < 2 {
		t.Error("expected jitter to produce varied delays")
	}
}

func TestRetryConfig_Sleep_RespectsContext(t *testing.T) {
	rc := RetryConfig{
		BaseDelay:   1 * time.Hour, // absurdly long
		MaxDelay:    1 * time.Hour,
		Multiplier:  1.0,
		JitterRatio: 0,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	start := time.Now()
	err := rc.Sleep(ctx, 0)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("expected error from cancelled context")
	}
	if elapsed > 100*time.Millisecond {
		t.Error("Sleep did not return promptly on cancelled context")
	}
}

func TestRetryConfig_Sleep_Completes(t *testing.T) {
	rc := RetryConfig{
		BaseDelay:   10 * time.Millisecond,
		MaxDelay:    100 * time.Millisecond,
		Multiplier:  1.0,
		JitterRatio: 0,
	}

	err := rc.Sleep(context.Background(), 0)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

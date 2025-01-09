package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/osmosis-labs/osmoutil-go/retry"
)

func TestRetryWithBackoff_Success(t *testing.T) {
	cfg := retry.RetryConfig{
		MaxDuration:       5 * time.Second,
		InitialInterval:   100 * time.Millisecond,
		MaxInterval:       500 * time.Millisecond,
		IntervalIncrement: 100 * time.Millisecond,
	}

	callCount := 0
	operation := func(ctx context.Context) error {
		callCount++
		return nil // Simulate a successful operation
	}

	err := retry.RetryWithBackoff(context.Background(), cfg, operation)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if callCount != 1 {
		t.Fatalf("expected operation to be called once, got %d", callCount)
	}
}

func TestRetryWithBackoff_Failure(t *testing.T) {
	cfg := retry.RetryConfig{
		MaxDuration:       1 * time.Second,
		InitialInterval:   100 * time.Millisecond,
		MaxInterval:       500 * time.Millisecond,
		IntervalIncrement: 100 * time.Millisecond,
	}

	callCount := 0
	operation := func(ctx context.Context) error {
		callCount++
		return errors.New("operation failed") // Simulate a failing operation
	}

	err := retry.RetryWithBackoff(context.Background(), cfg, operation)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	if callCount <= 1 {
		t.Fatalf("expected operation to be retried, got %d attempts", callCount)
	}
}

func TestRetryWithBackoff_MaxDuration(t *testing.T) {
	cfg := retry.RetryConfig{
		MaxDuration:       200 * time.Millisecond, // Set a short max duration
		InitialInterval:   50 * time.Millisecond,
		MaxInterval:       100 * time.Millisecond,
		IntervalIncrement: 10 * time.Millisecond,
	}

	callCount := 0
	operation := func(ctx context.Context) error {
		callCount++
		return errors.New("operation failed") // Simulate a failing operation
	}

	startTime := time.Now()
	err := retry.RetryWithBackoff(context.Background(), cfg, operation)
	duration := time.Since(startTime)

	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	if duration > cfg.MaxDuration+50*time.Millisecond { // Allow a small buffer for timing inaccuracies
		t.Fatalf("expected duration to be less than or equal to %v, got %v", cfg.MaxDuration, duration)
	}

	if callCount < 2 {
		t.Fatalf("expected operation to be retried at least twice, got %d attempts", callCount)
	}
}

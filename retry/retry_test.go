package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/osmosis-labs/osmoutil-go/retry"
	"github.com/stretchr/testify/assert"
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
	assert.NoError(t, err, "expected no error")

	assert.Equal(t, 1, callCount, "expected operation to be called once")
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
	assert.Error(t, err, "expected an error")

	assert.Greater(t, callCount, 1, "expected operation to be retried")
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

	assert.Error(t, err, "expected an error")

	assert.LessOrEqual(t, duration, cfg.MaxDuration+50*time.Millisecond, "expected duration to be within limit")

	assert.GreaterOrEqual(t, callCount, 2, "expected operation to be retried at least twice")
}

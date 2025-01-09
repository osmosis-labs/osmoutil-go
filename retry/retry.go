package retry

import (
	"context"
	"fmt"
	"time"
)

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	// MaxDuration is the maximum duration for the entire retry operation
	MaxDuration time.Duration
	// InitialInterval is the initial interval to retry the operation
	InitialInterval time.Duration
	// MaxInterval is the cap for the interval to retry the operation, as it grows linearly using IntervalIncrement
	MaxInterval time.Duration
	// IntervalIncrement is the increment interval to retry the operation
	IntervalIncrement time.Duration
}

// RetryWithBackoff executes an operation with linear backoff and timeout
// Returns error from operation or context error if cancelled
func RetryWithBackoff(ctx context.Context, cfg RetryConfig, operation func(context.Context) error) error {
	timer := time.NewTimer(cfg.MaxDuration)
	defer timer.Stop()

	interval := cfg.InitialInterval

	for {
		if err := operation(ctx); err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				return fmt.Errorf("operation timed out after %v: %w", cfg.MaxDuration, err)
			case <-time.After(interval):
				// Increase interval for next iteration
				// Cap the interval at MaxInterval
				interval = min(interval+cfg.IntervalIncrement, cfg.MaxInterval)
				continue
			}
		}
		return nil
	}
}

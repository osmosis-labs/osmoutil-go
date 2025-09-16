package retry

import (
	"context"
	"fmt"
	"strings"
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
// Optional nonRetriablePatterns will cause immediate failure without retry if error contains any of these strings
func RetryWithBackoff(ctx context.Context, cfg RetryConfig, operation func(context.Context) error, nonRetriablePatterns ...string) error {
	timer := time.NewTimer(cfg.MaxDuration)
	defer timer.Stop()

	interval := cfg.InitialInterval

	for {
		if err := operation(ctx); err != nil {
			// Check if this is a non-retriable error
			if isNonRetriable(err, nonRetriablePatterns) {
				return err // Return immediately, don't retry
			}

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

// isNonRetriable checks if an error contains any of the non-retriable patterns
func isNonRetriable(err error, nonRetriablePatterns []string) bool {
	if err == nil || len(nonRetriablePatterns) == 0 {
		return false
	}

	errStr := strings.ToLower(err.Error())
	for _, pattern := range nonRetriablePatterns {
		if strings.Contains(errStr, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

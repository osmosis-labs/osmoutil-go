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

func TestRetryWithBackoff_NonRetriablePatterns(t *testing.T) {
	cfg := retry.RetryConfig{
		MaxDuration:       5 * time.Second,
		InitialInterval:   100 * time.Millisecond,
		MaxInterval:       500 * time.Millisecond,
		IntervalIncrement: 100 * time.Millisecond,
	}

	tests := []struct {
		name                 string
		errorMessage         string
		nonRetriablePatterns []string
		expectRetry          bool
		expectedCallCount    int
	}{
		{
			name:                 "account sequence mismatch - should not retry",
			errorMessage:         "account sequence mismatch, expected 5, got 3",
			nonRetriablePatterns: []string{"account sequence mismatch", "invalid signature"},
			expectRetry:          false,
			expectedCallCount:    1,
		},
		{
			name:                 "invalid signature - should not retry",
			errorMessage:         "invalid signature for account",
			nonRetriablePatterns: []string{"account sequence mismatch", "invalid signature"},
			expectRetry:          false,
			expectedCallCount:    1,
		},
		{
			name:                 "insufficient funds - should not retry",
			errorMessage:         "Insufficient funds: 100uosmo < 1000uosmo",
			nonRetriablePatterns: []string{"insufficient funds", "account sequence mismatch"},
			expectRetry:          false,
			expectedCallCount:    1,
		},
		{
			name:                 "case insensitive pattern matching",
			errorMessage:         "ACCOUNT SEQUENCE MISMATCH detected",
			nonRetriablePatterns: []string{"account sequence mismatch"},
			expectRetry:          false,
			expectedCallCount:    1,
		},
		{
			name:                 "retriable error - should retry",
			errorMessage:         "network timeout occurred",
			nonRetriablePatterns: []string{"account sequence mismatch", "invalid signature"},
			expectRetry:          true,
			expectedCallCount:    -1, // Will retry multiple times within short duration, so we just check its greater than 1
		},
		{
			name:                 "no patterns provided - should retry",
			errorMessage:         "account sequence mismatch",
			nonRetriablePatterns: []string{},
			expectRetry:          true,
			expectedCallCount:    -1, // Will retry multiple times within short duration, so we just check its greater than 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			operation := func(ctx context.Context) error {
				callCount++
				return errors.New(tt.errorMessage)
			}

			// Use shorter max duration for retriable errors to speed up test
			testCfg := cfg
			if tt.expectRetry {
				testCfg.MaxDuration = 300 * time.Millisecond
			}

			err := retry.RetryWithBackoff(context.Background(), testCfg, operation, tt.nonRetriablePatterns...)

			assert.Error(t, err, "expected an error")

			if tt.expectRetry {
				assert.Greater(t, callCount, 1, "expected operation to be retried multiple times")
			} else {
				assert.Equal(t, tt.expectedCallCount, callCount, "expected operation to be called exactly once for non-retriable error")
			}
		})
	}
}

func TestIsNonRetriable(t *testing.T) {
	tests := []struct {
		name                 string
		err                  error
		nonRetriablePatterns []string
		expected             bool
	}{
		{
			name:                 "nil error",
			err:                  nil,
			nonRetriablePatterns: []string{"account sequence mismatch"},
			expected:             false,
		},
		{
			name:                 "empty patterns",
			err:                  errors.New("account sequence mismatch"),
			nonRetriablePatterns: []string{},
			expected:             false,
		},
		{
			name:                 "nil patterns",
			err:                  errors.New("account sequence mismatch"),
			nonRetriablePatterns: nil,
			expected:             false,
		},
		{
			name:                 "exact match - account sequence mismatch",
			err:                  errors.New("account sequence mismatch, expected 5, got 3"),
			nonRetriablePatterns: []string{"account sequence mismatch"},
			expected:             true,
		},
		{
			name:                 "exact match - invalid signature",
			err:                  errors.New("invalid signature for account"),
			nonRetriablePatterns: []string{"invalid signature"},
			expected:             true,
		},
		{
			name:                 "exact match - insufficient funds",
			err:                  errors.New("insufficient funds: 100uosmo < 1000uosmo"),
			nonRetriablePatterns: []string{"insufficient funds"},
			expected:             true,
		},
		{
			name:                 "case insensitive match - lowercase pattern",
			err:                  errors.New("ACCOUNT SEQUENCE MISMATCH detected"),
			nonRetriablePatterns: []string{"account sequence mismatch"},
			expected:             true,
		},
		{
			name:                 "case insensitive match - uppercase pattern",
			err:                  errors.New("account sequence mismatch detected"),
			nonRetriablePatterns: []string{"ACCOUNT SEQUENCE MISMATCH"},
			expected:             true,
		},
		{
			name:                 "case insensitive match - mixed case",
			err:                  errors.New("Account Sequence Mismatch detected"),
			nonRetriablePatterns: []string{"account SEQUENCE mismatch"},
			expected:             true,
		},
		{
			name:                 "partial match within error message",
			err:                  errors.New("error: account sequence mismatch occurred during transaction"),
			nonRetriablePatterns: []string{"account sequence mismatch"},
			expected:             true,
		},
		{
			name:                 "multiple patterns - first matches",
			err:                  errors.New("account sequence mismatch detected"),
			nonRetriablePatterns: []string{"account sequence mismatch", "invalid signature", "insufficient funds"},
			expected:             true,
		},
		{
			name:                 "multiple patterns - second matches",
			err:                  errors.New("invalid signature for account"),
			nonRetriablePatterns: []string{"account sequence mismatch", "invalid signature", "insufficient funds"},
			expected:             true,
		},
		{
			name:                 "multiple patterns - third matches",
			err:                  errors.New("insufficient funds: balance too low"),
			nonRetriablePatterns: []string{"account sequence mismatch", "invalid signature", "insufficient funds"},
			expected:             true,
		},
		{
			name:                 "no match - different error",
			err:                  errors.New("network timeout occurred"),
			nonRetriablePatterns: []string{"account sequence mismatch", "invalid signature"},
			expected:             false,
		},
		{
			name:                 "no match - partial string not found",
			err:                  errors.New("account sequence"),
			nonRetriablePatterns: []string{"account sequence mismatch"},
			expected:             false,
		},
		{
			name:                 "substring match but pattern is longer",
			err:                  errors.New("account"),
			nonRetriablePatterns: []string{"account sequence mismatch"},
			expected:             false,
		},
		{
			name:                 "common blockchain errors",
			err:                  errors.New("tx already exists in cache"),
			nonRetriablePatterns: []string{"tx already exists", "account sequence mismatch", "out of gas"},
			expected:             true,
		},
		{
			name:                 "gas related error",
			err:                  errors.New("out of gas in location: ReadPerByte"),
			nonRetriablePatterns: []string{"out of gas"},
			expected:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := retry.IsNonRetriable(tt.err, tt.nonRetriablePatterns)
			assert.Equal(t, tt.expected, result, "IsNonRetriable result mismatch")
		})
	}
}

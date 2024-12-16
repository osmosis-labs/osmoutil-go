package tx_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/osmosis-labs/osmoutil-go/tx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultForceRefetchInterval = time.Minute
	defaultTimeout              = time.Minute
)

func TestNonceTracker_IncrementAndGet(t *testing.T) {
	initialNonce := uint64(10)

	tracker := tx.NewNonceTracker(
		func(ctx context.Context) (tx.NonceResponse, error) {
			return tx.NonceResponse{Nonce: initialNonce, Accnum: 1}, nil
		}, defaultForceRefetchInterval, defaultTimeout)
	_, err := tracker.ForceRefetch(context.Background())
	require.NoError(t, err)

	// Note: first nonce does not get incremented.
	for i := 0; i <= 5; i++ {
		result := tracker.IncrementAndGet()
		assert.Equal(t, initialNonce+uint64(i), result.Nonce)
		assert.Equal(t, uint64(1), result.Accnum)
	}
}

func TestNonceTracker_ForceRefetch(t *testing.T) {
	tests := []struct {
		name                 string
		fetchNonceFunc       func(ctx context.Context) (tx.NonceResponse, error)
		initialLastRefetch   time.Time
		forceRefetchInterval time.Duration
		refetchTimeout       time.Duration
		shouldPrefetch       bool
		expectedError        bool
		expectedNonce        uint64
	}{
		{
			name: "Successful refetch",
			fetchNonceFunc: func(ctx context.Context) (tx.NonceResponse, error) {
				return tx.NonceResponse{Nonce: 20, Accnum: 1}, nil
			},
			initialLastRefetch:   time.Now().Add(-2 * time.Minute),
			forceRefetchInterval: time.Minute,
			refetchTimeout:       time.Second,
			expectedError:        false,
			expectedNonce:        20,
		},
		{
			name: "Too soon for refetch",
			fetchNonceFunc: func(ctx context.Context) (tx.NonceResponse, error) {
				return tx.NonceResponse{}, nil
			},
			initialLastRefetch:   time.Now(),
			forceRefetchInterval: time.Minute,
			refetchTimeout:       time.Second,
			shouldPrefetch:       true,
			expectedError:        true,
			expectedNonce:        0,
		},
		{
			name: "Fetch timeout",
			fetchNonceFunc: func(ctx context.Context) (tx.NonceResponse, error) {
				time.Sleep(2 * time.Second)
				return tx.NonceResponse{}, nil
			},
			initialLastRefetch:   time.Now().Add(-2 * time.Minute),
			forceRefetchInterval: time.Minute,
			refetchTimeout:       time.Second,
			expectedError:        true,
			expectedNonce:        0,
		},
		{
			name: "Fetch error",
			fetchNonceFunc: func(ctx context.Context) (tx.NonceResponse, error) {
				return tx.NonceResponse{}, errors.New("fetch error")
			},
			initialLastRefetch:   time.Now().Add(-2 * time.Minute),
			forceRefetchInterval: time.Minute,
			refetchTimeout:       time.Second,
			expectedError:        true,
			expectedNonce:        0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := tx.NewNonceTracker(tt.fetchNonceFunc, tt.forceRefetchInterval, tt.refetchTimeout)

			if tt.shouldPrefetch {
				_, err := tracker.ForceRefetch(context.Background())
				require.NoError(t, err)
			}

			result, err := tracker.ForceRefetch(context.Background())

			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, uint64(0), result.Nonce)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedNonce, result.Nonce)
				assert.True(t, tracker.GetLastRefetchTime().After(tt.initialLastRefetch))
			}
		})
	}
}

func TestNewNonceTracker(t *testing.T) {
	fetchNonce := func(ctx context.Context) (tx.NonceResponse, error) {
		return tx.NonceResponse{Nonce: 1, Accnum: 1}, nil
	}

	tracker := tx.NewNonceTracker(fetchNonce, defaultForceRefetchInterval, defaultTimeout)

	assert.NotNil(t, tracker)
	assert.True(t, tracker.GetLastRefetchTime().IsZero())
	assert.Equal(t, tracker.IncrementAndGet(), tx.NonceResponse{})
}

package tx

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// NonceTrackerI is an interface defining nonce tracking.
type NonceTrackerI interface {
	// IncrementAndGet increments the internal nonce by one under lock, returning
	// the updated value. Account number does not change.
	// If this is the first fetch, does not increment.
	IncrementAndGet() NonceResponse
	// ForceRefetch refetches the nonce under a lock.
	// Refetching happens under a pre-configured timeout to avoid deadlock.
	// Returns error if already force refetched within the a pre-configured interval.
	// The interval logic exists to avoid several concurrent clients attempting to refetch
	// concurrently.
	ForceRefetch(ctx context.Context) (NonceResponse, error)

	// GetLastRefetchTime returns the time of the last refetch.
	GetLastRefetchTime() time.Time
}

type NonceTracker struct {
	nonceData  NonceResponse
	mu         sync.RWMutex
	fetchNonce func(ctx context.Context) (NonceResponse, error)

	cancelCh             chan struct{}
	forceRefetchInterval time.Duration
	refetchTimeout       time.Duration
	isFirstFetch         bool

	lastRefetch time.Time
}

// NonceResponse contains nonce/sequence number and
// an account number (optional)
//
// Cosmos accounts have account number. Ethereum do not.
type NonceResponse struct {
	Nonce  uint64
	Accnum uint64
}

var _ NonceTrackerI = &NonceTracker{}

var UnsetNonceTracker NonceTrackerI = nil

// NewNonceTracker returns a new instance of nonce tracker with the given parameters.
// It does not pre-fetch the nonce. The caller must call ForceRefetch(...)
func NewNonceTracker(fetchNonce func(ctx context.Context) (NonceResponse, error), forceRefetchInterval time.Duration, refetchTimeout time.Duration) *NonceTracker {
	return &NonceTracker{
		cancelCh:             make(chan struct{}),
		fetchNonce:           fetchNonce,
		forceRefetchInterval: forceRefetchInterval,
		refetchTimeout:       refetchTimeout,
		isFirstFetch:         true,
	}
}

// NewNonceTrackerWithRefetch initializes a new nonce tracker and executes ForceRefetch().
func NewNonceTrackerWithRefetch(ctx context.Context, fetchNonce func(ctx context.Context) (NonceResponse, error), forceRefetchInterval time.Duration, refetchTimeout time.Duration) (*NonceTracker, error) {
	// Initialize the nonce tracker
	nonceTracker := NewNonceTracker(fetchNonce, forceRefetchInterval, refetchTimeout)

	// Force refetch to get the initial nonce
	if _, err := nonceTracker.ForceRefetch(ctx); err != nil {
		return nil, fmt.Errorf("failed to force refetch nonce tracker: %v", err)
	}

	return nonceTracker, nil
}

// GetLastRefetchTime implements NonceTrackerI.
func (n *NonceTracker) GetLastRefetchTime() time.Time {
	return n.lastRefetch
}

// IncrementAndGet implements NonceTrackerI
func (n *NonceTracker) IncrementAndGet() NonceResponse {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Increment only on any fetch after the first one.
	if !n.isFirstFetch {
		n.nonceData.Nonce++
	} else {
		n.isFirstFetch = false
	}

	result := n.nonceData

	return result
}

// ForceRefetch implements NonceTrackerI
func (n *NonceTracker) ForceRefetch(ctx context.Context) (NonceResponse, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	timeSince := time.Since(n.lastRefetch)
	if timeSince > n.forceRefetchInterval {
		return n.refetchAndUpdateNonce(ctx)
	}

	return NonceResponse{}, fmt.Errorf("failed to force refetch time since (%s), force refetch interval (%s)", timeSince, n.forceRefetchInterval)
}

// refetchAndUpdateNonce refetched and updates internal nonce.
// Returns error if already force refetched within the a pre-configured interval.
// Returns error if refetching times out per internal configuration.
// Returns updated nonce data on success.
// Updates last refetch time on success.s
// CONTRACT: called handles concurrency
func (n *NonceTracker) refetchAndUpdateNonce(ctx context.Context) (NonceResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, n.refetchTimeout)
	defer cancel()

	type result struct {
		nonce NonceResponse
		err   error
	}

	ch := make(chan result, 1)

	go func(ctx context.Context) {
		nonce, err := n.fetchNonce(ctx)
		ch <- result{nonce: nonce, err: err}
	}(ctx)

	select {
	case <-ctx.Done():
		return NonceResponse{}, ctx.Err()
	case res := <-ch:
		if res.err != nil {
			return NonceResponse{}, res.err
		}
		n.nonceData = res.nonce
		n.lastRefetch = time.Now()

		return n.nonceData, nil
	}
}

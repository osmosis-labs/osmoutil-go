package mocks

import (
	"context"
	"time"

	"github.com/osmosis-labs/osmoutil-go/tx"
)

type NonceTrackerMock struct {
	ForceRefetchFunc       func(ctx context.Context) (tx.NonceResponse, error)
	GetLastRefetchTimeFunc func() time.Time
	IncrementAndGetFunc    func() tx.NonceResponse
}

// ForceUpdateNonce implements tx.NonceTrackerI.
func (n *NonceTrackerMock) ForceUpdateNonce(nonce uint64) {
	panic("unimplemented")
}

// GetCurrentNonce implements tx.NonceTrackerI.
func (n *NonceTrackerMock) GetCurrentNonce() tx.NonceResponse {
	panic("unimplemented")
}

// ForceRefetch implements tx.NonceTrackerI.
func (n *NonceTrackerMock) ForceRefetch(ctx context.Context) (tx.NonceResponse, error) {
	if n.ForceRefetchFunc == nil {
		panic("unimplemented")
	}
	return n.ForceRefetchFunc(ctx)
}

// GetLastRefetchTime implements tx.NonceTrackerI.
func (n *NonceTrackerMock) GetLastRefetchTime() time.Time {
	if n.GetLastRefetchTimeFunc == nil {
		panic("unimplemented")
	}
	return n.GetLastRefetchTimeFunc()
}

// IncrementAndGet implements tx.NonceTrackerI.
func (n *NonceTrackerMock) IncrementAndGet() tx.NonceResponse {
	if n.IncrementAndGetFunc == nil {
		panic("unimplemented")
	}
	return n.IncrementAndGetFunc()
}

var _ tx.NonceTrackerI = &NonceTrackerMock{}

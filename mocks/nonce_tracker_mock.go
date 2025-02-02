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

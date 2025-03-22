package mocks

import (
	"context"

	"github.com/cosmos/cosmos-sdk/types/tx"
	broadcastcosmos "github.com/osmosis-labs/osmoutil-go/tx/broadcast/cosmos"
)

type MockCosmosRestClient struct {
	GetUrlFunc             func() string
	GetInitialSequenceFunc func(ctx context.Context, address string) (uint64, uint64, error)
	GetAllBalancesFunc     func(ctx context.Context, address string) (broadcastcosmos.BalancesResponse, error)
	SimulateGasUsedFunc    func(ctx context.Context, simulateReq *tx.SimulateRequest) (uint64, error)
}

func (m *MockCosmosRestClient) GetUrl() string {
	if m.GetUrlFunc != nil {
		return m.GetUrlFunc()
	}
	return ""
}

func (m *MockCosmosRestClient) GetInitialSequence(ctx context.Context, address string) (uint64, uint64, error) {
	if m.GetInitialSequenceFunc != nil {
		return m.GetInitialSequenceFunc(ctx, address)
	}
	return 0, 0, nil
}

func (m *MockCosmosRestClient) GetAllBalances(ctx context.Context, address string) (broadcastcosmos.BalancesResponse, error) {
	if m.GetAllBalancesFunc != nil {
		return m.GetAllBalancesFunc(ctx, address)
	}
	return broadcastcosmos.BalancesResponse{}, nil
}

func (m *MockCosmosRestClient) SimulateGasUsed(ctx context.Context, simulateReq *tx.SimulateRequest) (uint64, error) {
	if m.SimulateGasUsedFunc != nil {
		return m.SimulateGasUsedFunc(ctx, simulateReq)
	}
	return 0, nil
}

var _ broadcastcosmos.CosmosRESTClient = &MockCosmosRestClient{}

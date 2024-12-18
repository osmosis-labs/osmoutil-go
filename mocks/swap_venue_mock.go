package mocks

import (
	"context"

	swapvenuetypes "github.com/osmosis-labs/osmoutil-go/swapvenue/types"
)

type MockSwapVenue struct {
	GetBalanceFunc              func(ctx context.Context, denom string) (float64, error)
	GetBalancesFunc             func(ctx context.Context, denoms ...string) (map[string]float64, error)
	GetNameFunc                 func() string
	GetPriceFunc                func(ctx context.Context, pair swapvenuetypes.SwapVenuePairI) (float64, error)
	GetSwapVenuePairsFunc       func(pair swapvenuetypes.AbstractSwapPair) []swapvenuetypes.SwapVenuePairI
	GetTradingFeeFunc           func() float64
	MarketBuyFunc               func(ctx context.Context, pair swapvenuetypes.SwapVenuePairI, amount float64) (swapvenuetypes.OrderResult, error)
	MarketSellFunc              func(ctx context.Context, pair swapvenuetypes.SwapVenuePairI, amount float64) (swapvenuetypes.OrderResult, error)
	RegisterSupportedAssetsFunc func(assets []swapvenuetypes.AssetI)
	RegisterSwapVenuePairFunc   func(pair swapvenuetypes.AbstractSwapPair, venuePairs []swapvenuetypes.SwapVenuePairI)
}

// GetBalance implements swapvenuetypes.SwapVenueI.
func (m *MockSwapVenue) GetBalance(ctx context.Context, denom string) (float64, error) {
	if m.GetBalanceFunc != nil {
		return m.GetBalanceFunc(ctx, denom)
	}
	return 0, nil
}

// GetBalances implements swapvenuetypes.SwapVenueI.
func (m *MockSwapVenue) GetBalances(ctx context.Context, denoms ...string) (map[string]float64, error) {
	if m.GetBalancesFunc != nil {
		return m.GetBalancesFunc(ctx, denoms...)
	}
	return nil, nil
}

// GetName implements swapvenuetypes.SwapVenueI.
func (m *MockSwapVenue) GetName() string {
	if m.GetNameFunc != nil {
		return m.GetNameFunc()
	}
	return ""
}

// GetPrice implements swapvenuetypes.SwapVenueI.
func (m *MockSwapVenue) GetPrice(ctx context.Context, pair swapvenuetypes.SwapVenuePairI) (float64, error) {
	if m.GetPriceFunc != nil {
		return m.GetPriceFunc(ctx, pair)
	}
	return 0, nil
}

// GetSwapVenuePairs implements swapvenuetypes.SwapVenueI.
func (m *MockSwapVenue) GetSwapVenuePairs(pair swapvenuetypes.AbstractSwapPair) []swapvenuetypes.SwapVenuePairI {
	if m.GetSwapVenuePairsFunc != nil {
		return m.GetSwapVenuePairsFunc(pair)
	}
	return nil
}

// GetTradingFee implements swapvenuetypes.SwapVenueI.
func (m *MockSwapVenue) GetTradingFee() float64 {
	if m.GetTradingFeeFunc != nil {
		return m.GetTradingFeeFunc()
	}
	return 0
}

// MarketBuy implements swapvenuetypes.SwapVenueI.
func (m *MockSwapVenue) MarketBuy(ctx context.Context, pair swapvenuetypes.SwapVenuePairI, amount float64) (swapvenuetypes.OrderResult, error) {
	if m.MarketBuyFunc != nil {
		return m.MarketBuyFunc(ctx, pair, amount)
	}
	return swapvenuetypes.OrderResult{}, nil
}

// MarketSell implements swapvenuetypes.SwapVenueI.
func (m *MockSwapVenue) MarketSell(ctx context.Context, pair swapvenuetypes.SwapVenuePairI, amount float64) (swapvenuetypes.OrderResult, error) {
	if m.MarketSellFunc != nil {
		return m.MarketSellFunc(ctx, pair, amount)
	}
	return swapvenuetypes.OrderResult{}, nil
}

// RegisterSupportedAssets implements swapvenuetypes.SwapVenueI.
func (m *MockSwapVenue) RegisterSupportedAssets(assets []swapvenuetypes.AssetI) {
	if m.RegisterSupportedAssetsFunc != nil {
		m.RegisterSupportedAssetsFunc(assets)
	}
}

// RegisterSwapVenuePair implements swapvenuetypes.SwapVenueI.
func (m *MockSwapVenue) RegisterSwapVenuePair(pair swapvenuetypes.AbstractSwapPair, venuePairs []swapvenuetypes.SwapVenuePairI) {
	if m.RegisterSwapVenuePairFunc != nil {
		m.RegisterSwapVenuePairFunc(pair, venuePairs)
	}
}

var _ swapvenuetypes.SwapVenueI = &MockSwapVenue{}

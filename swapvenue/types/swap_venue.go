package swapvenuetypes

import "context"

// SwapVenueI is the interface for a swap venue
type SwapVenueI interface {
	// GetName returns the name of the venue
	GetName() string

	// GetPrice returns normalized price of the pair (exponents applied).
	GetPrice(ctx context.Context, pair SwapVenuePairI) (float64, error)

	// MarketBuy buys the amount of the pair at the current market price.
	// CONTRACT: the asset exponents are applied to the amounts.
	MarketBuy(ctx context.Context, pair SwapVenuePairI, amount float64) (OrderResult, error)

	// MarketSell sells the amount of the pair at the current market price.
	// CONTRACT: the asset exponents are applied to the amounts.
	MarketSell(ctx context.Context, pair SwapVenuePairI, amount float64) (OrderResult, error)

	// GetBalance returns normalized balance (exponents applied)
	GetBalance(ctx context.Context, denom string) (float64, error)

	// GetBalances returns normalized balances (exponents applied) for the given denoms.
	// CONTRACT: the asset exponents are applied to the amounts.
	GetBalances(ctx context.Context, denoms []string) (map[string]float64, error)

	// GetTradingFee returns the trading fee for the venue.
	GetTradingFee() float64

	// GetSwapVenuePairs returns the venue-native pairs supported by the venue
	// given an abstract pair.
	GetSwapVenuePairs(pair AbstractSwapPair) []SwapVenuePairI

	// RegisterSwapVenuePair registers the pairs supported by the venue.
	RegisterSwapVenuePair(pair AbstractSwapPair, venuePairs []SwapVenuePairI)

	// RegisterSupportedAssets registers the assets supported by the venue.
	RegisterSupportedAssets(assets []AssetI)
}

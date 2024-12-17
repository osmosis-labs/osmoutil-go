package swapvenuetypes

// AbstractSwapPair is the interface for an abstract swap pair.
// For example, a pair of BTC/USDT is an abstract pair.
// In contrast, a pair of ALLBTC/ALLUSDT on Osmosis is a venue-native pair.
type AbstractSwapPair struct {
	PreferredBuyVenue string
	Base              string
	Quote             string
}

package binance

// Returns a concrete implementation of the BinanceSwapVenue.
func NewBinanceSwapVenueConcrete(config BinanceSwapVenueConfig) *BinanceSwapVenue {
	return newBinanceSwapVenue(config)
}

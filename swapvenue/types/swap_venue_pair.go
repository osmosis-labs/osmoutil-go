package swapvenuetypes

// SwapVenuePairI is the interface for a swap venue pair.
type SwapVenuePairI interface {
	GetBase() AssetI
	GetQuote() AssetI
	GetMinAmount() float64
	GetMaxAmount() float64
}

// OrderResult is the result of a swap venue order.
type OrderResult struct {
	// QuoteAmount is the amount of the quote asset.
	QuoteAmount float64
	// Price is the price of the order.
	Price float64
	// TradeID is the ID of the trade.
	TradeID string
}

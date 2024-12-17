package binance

// binancePriceResponse is the response type for the Binance price endpoint.
type binancePriceResponse struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

package binance

import (
	swapvenuetypes "github.com/osmosis-labs/osmoutil-go/swapvenue/types"
)

// BinanceSwapPair is a swap pair for Binance.
type BinanceSwapPair struct {
	// Base is the base asset of the swap pair.
	Base swapvenuetypes.AssetI
	// Quote is the quote asset of the swap pair.
	Quote swapvenuetypes.AssetI
	// MinAmount is the minimum amount of the swap pair.
	MinAmount float64
	// MaxAmount is the maximum amount of the swap pair.
	MaxAmount float64
}

// GetBase implements domain.SwapVenuePairI.
func (b *BinanceSwapPair) GetBase() swapvenuetypes.AssetI {
	return b.Base
}

// GetMaxAmount implements domain.SwapVenuePairI.
func (b *BinanceSwapPair) GetMaxAmount() float64 {
	return b.MaxAmount
}

// GetMinAmount implements domain.SwapVenuePairI.
func (b *BinanceSwapPair) GetMinAmount() float64 {
	return b.MinAmount
}

// GetQuote implements domain.SwapVenuePairI.
func (b *BinanceSwapPair) GetQuote() swapvenuetypes.AssetI {
	return b.Quote
}

// NewBinanceSwapPair returns a new BinanceSwapPair.
func NewBinanceSwapPair(base swapvenuetypes.AssetI, quote swapvenuetypes.AssetI, minAmount float64, maxAmount float64) *BinanceSwapPair {
	return &BinanceSwapPair{
		Base:      base,
		Quote:     quote,
		MinAmount: minAmount,
		MaxAmount: maxAmount,
	}
}

var _ swapvenuetypes.SwapVenuePairI = (*BinanceSwapPair)(nil)

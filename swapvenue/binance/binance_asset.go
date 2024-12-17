package binance

import swapvenuetypes "github.com/osmosis-labs/osmoutil-go/swapvenue/types"

// BinanceAsset is an asset for Binance.
type BinanceAsset struct {
	// Symbol is the symbol of the asset.
	Symbol string `json:"symbol"`
	// Name is the name of the asset.
	Name string `json:"name"`
}

// GetDenom implements domain.AssetI.
func (b *BinanceAsset) GetDenom() string {
	return b.Symbol
}

var _ swapvenuetypes.AssetI = (*BinanceAsset)(nil)

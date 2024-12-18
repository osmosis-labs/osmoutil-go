package binance

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/adshao/go-binance/v2"
	"github.com/osmosis-labs/osmoutil-go/httputil"
	swapvenuetypes "github.com/osmosis-labs/osmoutil-go/swapvenue/types"
)

// BinanceSwapVenue is a swap venue for Binance.
type BinanceSwapVenue struct {
	assets         []swapvenuetypes.AssetI
	swapVenuePairs map[swapvenuetypes.AbstractSwapPair][]swapvenuetypes.SwapVenuePairI

	config BinanceSwapVenueConfig
}

const (
	BinanceVenueName = "binance"

	DefaultBinanceURL = "https://api.binance.com/api/v3"
)

// BinanceSwapVenueConfig is the configuration for the BinanceSwapVenue.
type BinanceSwapVenueConfig struct {
	// URL is the URL of the Binance API.
	URL string
	// APIKey is the API key for the Binance API.
	APIKey string
	// SecretKey is the secret key for the Binance API.
	SecretKey string
}

func NewBinanceSwapVenue(config BinanceSwapVenueConfig) swapvenuetypes.SwapVenueI {
	return newBinanceSwapVenue(config)
}

func newBinanceSwapVenue(config BinanceSwapVenueConfig) *BinanceSwapVenue {
	return &BinanceSwapVenue{
		assets:         make([]swapvenuetypes.AssetI, 0),
		swapVenuePairs: make(map[swapvenuetypes.AbstractSwapPair][]swapvenuetypes.SwapVenuePairI),
		config:         config,
	}
}

// MarketBuy implements domain.SwapVenueI.
func (b *BinanceSwapVenue) MarketBuy(ctx context.Context, pair swapvenuetypes.SwapVenuePairI, amount float64) (swapvenuetypes.OrderResult, error) {
	client := binance.NewClient(b.config.APIKey, b.config.SecretKey)

	amountStr := strconv.FormatFloat(amount, 'f', -1, 64)

	baseQuote := formatBaseQuote(pair)

	order, err := client.NewCreateOrderService().Symbol(baseQuote).Side(binance.SideTypeBuy).Type(binance.OrderTypeMarket).Quantity(amountStr).Do(ctx)
	if err != nil {
		return swapvenuetypes.OrderResult{}, err
	}

	boughtPrice, err := strconv.ParseFloat(order.Fills[0].Price, 64)
	if err != nil {
		return swapvenuetypes.OrderResult{}, err
	}

	boughtAmount, err := strconv.ParseFloat(order.ExecutedQuantity, 64)
	if err != nil {
		return swapvenuetypes.OrderResult{}, err
	}

	return swapvenuetypes.OrderResult{
		QuoteAmount: boughtAmount,
		Price:       boughtPrice,
	}, nil
}

// GetBalance implements domain.SwapVenueI.
func (b *BinanceSwapVenue) GetBalance(ctx context.Context, denom string) (float64, error) {
	balances, err := b.GetBalances(ctx, denom)
	if err != nil {
		return 0, err
	}

	return balances[denom], nil
}

// GetBalances implements domain.SwapVenueI.
func (b *BinanceSwapVenue) GetBalances(ctx context.Context, denoms ...string) (map[string]float64, error) {
	client := binance.NewClient(b.config.APIKey, b.config.SecretKey)
	accountService := client.NewGetAccountService().OmitZeroBalances(true)

	// Get account snapshot
	res, err := accountService.Do(ctx)
	if err != nil {
		return nil, err
	}

	includeAll := len(denoms) == 0

	balances := make(map[string]float64)
	for _, balance := range res.Balances {
		if slices.Contains(denoms, balance.Asset) || includeAll {

			// Parse float
			parsedBalance, err := strconv.ParseFloat(balance.Free, 64)
			if err != nil {
				return nil, err
			}

			balances[balance.Asset] = parsedBalance
		}
	}

	return balances, nil
}

// GetPrice implements domain.SwapVenueI.
func (b *BinanceSwapVenue) GetPrice(ctx context.Context, pair swapvenuetypes.SwapVenuePairI) (float64, error) {
	baseQuote := formatBaseQuote(pair)

	url := fmt.Sprintf("%s/ticker/price?symbol=%s", b.config.URL, baseQuote)

	var binancePriceResponse binancePriceResponse
	if err := httputil.RunGet(ctx, url, nil, &binancePriceResponse); err != nil {
		return 0, err
	}

	priceFloat, err := strconv.ParseFloat(binancePriceResponse.Price, 10)
	if err != nil {
		return 0, err
	}

	return priceFloat, nil
}

// GetTradingFee implements domain.SwapVenueI.
func (b *BinanceSwapVenue) GetTradingFee() float64 {
	// TODO: set to something reasonable
	return 0
}

// MarketSell implements domain.SwapVenueI.
func (b *BinanceSwapVenue) MarketSell(ctx context.Context, pair swapvenuetypes.SwapVenuePairI, amount float64) (swapvenuetypes.OrderResult, error) {
	client := binance.NewClient(b.config.APIKey, b.config.SecretKey)

	amountStr := strconv.FormatFloat(amount, 'f', 8, 64)

	baseQuote := formatBaseQuote(pair)

	order, err := client.NewCreateOrderService().Symbol(baseQuote).Side(binance.SideTypeSell).Type(binance.OrderTypeMarket).Quantity(amountStr).Do(ctx)
	if err != nil {
		return swapvenuetypes.OrderResult{}, err
	}

	soldPrice, err := strconv.ParseFloat(order.Fills[0].Price, 64)
	if err != nil {
		return swapvenuetypes.OrderResult{}, err
	}

	soldAmount, err := strconv.ParseFloat(order.CummulativeQuoteQuantity, 64)
	if err != nil {
		return swapvenuetypes.OrderResult{}, err
	}

	return swapvenuetypes.OrderResult{
		QuoteAmount: soldAmount,
		Price:       soldPrice,
		TradeID:     strconv.FormatInt(order.OrderID, 10),
	}, nil
}

// GetSwapVenuePairs implements domain.SwapVenueI.
func (b *BinanceSwapVenue) GetSwapVenuePairs(pair swapvenuetypes.AbstractSwapPair) []swapvenuetypes.SwapVenuePairI {
	return b.swapVenuePairs[pair]
}

// RegisterSupportedAssets implements domain.SwapVenueI.
func (b *BinanceSwapVenue) RegisterSupportedAssets(assets []swapvenuetypes.AssetI) {
	b.assets = append(b.assets, assets...)
}

// RegisterSwapVenuePair implements domain.SwapVenueI.
func (b *BinanceSwapVenue) RegisterSwapVenuePair(pair swapvenuetypes.AbstractSwapPair, venuePairs []swapvenuetypes.SwapVenuePairI) {
	if _, ok := b.swapVenuePairs[pair]; !ok {
		b.swapVenuePairs[pair] = venuePairs
	} else {
		b.swapVenuePairs[pair] = append(b.swapVenuePairs[pair], venuePairs...)
	}
}

func (b *BinanceSwapVenue) GetUserAssets(ctx context.Context) ([]swapvenuetypes.AssetI, error) {

	client := binance.NewClient(b.config.APIKey, b.config.SecretKey)

	assets, err := client.NewGetUserAsset().Asset("").Do(ctx)
	if err != nil {
		return nil, err
	}

	for _, asset := range assets {
		b.assets = append(b.assets, &BinanceAsset{Symbol: asset.Asset})
	}

	return b.assets, nil
}

func (b *BinanceSwapVenue) GetVenueAssets(ctx context.Context) ([]swapvenuetypes.AssetI, error) {

	client := binance.NewClient(b.config.APIKey, b.config.SecretKey)

	assets, err := client.NewGetAllCoinsInfoService().Do(ctx)
	if err != nil {
		return nil, err
	}

	for _, asset := range assets {
		b.assets = append(b.assets, &BinanceAsset{Symbol: asset.Coin, Name: asset.Name})
	}

	return b.assets, nil
}

func formatBaseQuote(pair swapvenuetypes.SwapVenuePairI) string {
	return fmt.Sprintf("%s%s", pair.GetBase().GetDenom(), pair.GetQuote().GetDenom())
}

// GetName implements domain.SwapVenueI.
func (b *BinanceSwapVenue) GetName() string {
	return BinanceVenueName
}

var _ swapvenuetypes.SwapVenueI = &BinanceSwapVenue{}

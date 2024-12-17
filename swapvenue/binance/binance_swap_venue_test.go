package binance_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/osmosis-labs/osmoutil-go/swapvenue/binance"
	"github.com/stretchr/testify/require"
)

var (
	defaultPar = binance.NewBinanceSwapPair(&binance.BinanceAsset{
		Symbol: "BTC",
	}, &binance.BinanceAsset{
		Symbol: "USDT",
	}, 0.00006, 0.001)
)

// setupConfig returns a BinanceSwapVenueConfig for testing.
// Note: set the config to your own keys.
func setupConfig() binance.BinanceSwapVenueConfig {
	return binance.BinanceSwapVenueConfig{
		URL: binance.DefaultBinanceURL,

		// Note: set the config to your own keys.
		APIKey:    os.Getenv("BINANCE_API_KEY"),
		SecretKey: os.Getenv("BINANCE_SECRET_KEY"),
	}
}

var config = setupConfig()

func TestBinanceSwapVenue_MarketBuy(t *testing.T) {

	t.Skip("skip integration test")

	binanceClient := binance.NewBinanceSwapVenue(config)

	ctx := context.Background()

	orderResult, err := binanceClient.MarketBuy(ctx, defaultPar, 0.00005)
	require.NoError(t, err)

	fmt.Println(orderResult)
}

func TestBinanceSwapVenue_MarketSell(t *testing.T) {

	t.Skip("skip integration test")

	binanceClient := binance.NewBinanceSwapVenue(config)

	ctx := context.Background()

	orderResult, err := binanceClient.MarketSell(ctx, defaultPar, 0.00005)
	require.NoError(t, err)

	fmt.Println(orderResult)
}
func TestBinanceGetPrice(t *testing.T) {

	t.Skip("skip integration test")

	binanceClient := binance.NewBinanceSwapVenue(config)

	ctx := context.Background()

	price, err := binanceClient.GetPrice(ctx, defaultPar)
	require.NoError(t, err)

	t.Log(price)
}

func TestBinanceSwapVenue_GetUserAssets(t *testing.T) {

	t.Skip("skip integration test")

	binanceClient := binance.NewBinanceSwapVenue(config)

	ctx := context.Background()

	assets, err := binanceClient.GetUserAssets(ctx)
	require.NoError(t, err)

	t.Log(assets)
}

func TestBinanceSwapVenue_GetVenueAssets(t *testing.T) {

	t.Skip("skip integration test")

	binanceClient := binance.NewBinanceSwapVenue(config)

	ctx := context.Background()

	assets, err := binanceClient.GetVenueAssets(ctx)
	require.NoError(t, err)

	t.Log(assets)
}

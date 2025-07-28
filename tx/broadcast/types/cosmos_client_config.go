package broadcasttypes

import "time"

type CosmosClientConfig struct {
	Name              string
	NativeChainID     string
	Bech32Prefix      string
	FeeTokenDenom     string
	FeeTokenPrecision int
	AverageGasPrice   string
	Memo              string
	RPCURL            string
	LCDURL            string

	// Custom force refetch interval and refetch timeout.
	// If not set, the default values will be used.
	// Custom intervals must be set together. Either both custom or none of them.
	ForceRefetchInterval time.Duration
	RefetchTimeout       time.Duration
}

var (
	OsmosisClientConfig = CosmosClientConfig{
		Name:              "osmosis",
		NativeChainID:     "osmosis-1",
		Bech32Prefix:      "osmo",
		FeeTokenDenom:     "uosmo",
		FeeTokenPrecision: 6,
		AverageGasPrice:   "0.025",
		RPCURL:            "https://rpc.osmosis.zone",
		LCDURL:            "https://lcd.osmosis.zone",
		Memo:              "",
	}
)

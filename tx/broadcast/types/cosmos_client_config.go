package broadcasttypes

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

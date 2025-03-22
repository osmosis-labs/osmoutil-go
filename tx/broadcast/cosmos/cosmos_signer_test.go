package broadcastcosmos_test

import (
	"context"
	"testing"

	"github.com/osmosis-labs/osmoutil-go/mocks"
	osmoutilmocks "github.com/osmosis-labs/osmoutil-go/mocks"

	broadcastcosmos "github.com/osmosis-labs/osmoutil-go/tx/broadcast/cosmos"
	broadcasttypes "github.com/osmosis-labs/osmoutil-go/tx/broadcast/types"

	"github.com/stretchr/testify/require"
)

const (
	expectedAddress = "osmo1xrlgm0yqs6yce9l88q8p9v5kdhps054f0nlytq"

	throwawayPK = "17cdde089548458029596063aa7742758757eafcccf7e8cc7071b09ab1fad9a5"
)

var (
	osmosisClientConfig = broadcasttypes.OsmosisClientConfig
)

func TestCosmosSigner_GetAddressString(t *testing.T) {
	t.Parallel()

	// Setup
	nonceTracker := osmoutilmocks.NonceTrackerMock{}

	signer, err := broadcastcosmos.NewCosmosSigner(throwawayPK, osmosisClientConfig.Bech32Prefix, osmosisClientConfig.NativeChainID, osmosisClientConfig.FeeTokenDenom)
	require.NoError(t, err)

	signer.SetNonceTracker(&nonceTracker)

	// System under test
	address := signer.GetAddressString()
	require.Equal(t, expectedAddress, address)
}

func TestCosmosSigner_InitializeCosmosSigner(t *testing.T) {
	t.Parallel()

	// Setup
	ctx := context.Background()

	restClient := &mocks.MockCosmosRestClient{
		GetInitialSequenceFunc: func(ctx context.Context, address string) (uint64, uint64, error) {
			return 0, 0, nil
		},
	}

	// System under test
	signer, err := broadcastcosmos.InitializeCosmosSigner(ctx, throwawayPK, osmosisClientConfig, restClient)
	require.NoError(t, err)

	// Assertions
	require.Equal(t, expectedAddress, signer.GetAddressString())
}

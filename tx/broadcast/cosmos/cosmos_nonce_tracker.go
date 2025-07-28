package broadcastcosmos

import (
	"context"
	"time"

	osmoutilstx "github.com/osmosis-labs/osmoutil-go/tx"
)

const (
	defaultForceRefetchInterval = time.Second * 5
	defaultForceRefetchTimeout  = time.Second * 10
)

// NewCosmosNonceTracker creates a new nonce tracker for Cosmos
func NewCosmosNonceTracker(bech32Address string, restClient CosmosRESTClient) *osmoutilstx.NonceTracker {
	// Create a wrapper function to convert the sequence response
	getNonce := func(ctx context.Context) (osmoutilstx.NonceResponse, error) {
		seq, accNum, err := restClient.GetInitialSequence(ctx, bech32Address)
		if err != nil {
			return osmoutilstx.NonceResponse{}, err
		}
		return osmoutilstx.NonceResponse{
			Nonce:  seq,
			Accnum: accNum,
		}, nil
	}

	return osmoutilstx.NewNonceTracker(
		getNonce,
		defaultForceRefetchInterval,
		defaultForceRefetchTimeout,
	)
}

package httputil_test

import (
	"context"
	"testing"

	"github.com/osmosis-labs/osmoutil-go/httputil"
	"github.com/stretchr/testify/require"
)

func TestRunGet(t *testing.T) {

	t.Skip("skipping integration test")

	ctx := context.Background()

	url := "https://sqs.osmosis.zone/tokens/prices?baseDenoms=uosmo&humanDenoms=false"

	err := httputil.RunGet(ctx, url, nil, nil)
	require.NoError(t, err)
}

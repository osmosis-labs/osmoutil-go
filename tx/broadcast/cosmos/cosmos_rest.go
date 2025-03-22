package broadcastcosmos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	tx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/osmosis-labs/osmoutil-go/httputil"
)

// CosmosRESTClient is an interface for the Cosmos REST client
type CosmosRESTClient interface {
	// GetUrl returns the REST endpoint URL
	GetUrl() string

	// GetInitialSequence returns the initial sequence and account number
	GetInitialSequence(ctx context.Context, address string) (uint64, uint64, error)

	// GetAllBalances returns all balances for an address
	GetAllBalances(ctx context.Context, address string) (BalancesResponse, error)

	// SimulateGasUsed simulates a transaction to estimate gas usage
	SimulateGasUsed(ctx context.Context, simulateReq *tx.SimulateRequest) (uint64, error)
}

// CosmosRestClient provides a base implementation of the RestClient interface
type cosmosRestClient struct {
	url string
}

// NewCosmosRestClient creates a new CosmosRestClient instance
func NewCosmosRestClient(url string) (*cosmosRestClient, error) {
	if err := validateUrl(url); err != nil {
		return nil, fmt.Errorf("invalid REST URL: %w", err)
	}

	return &cosmosRestClient{
		url: url,
	}, nil
}

// GetUrl returns the REST endpoint URL
func (c *cosmosRestClient) GetUrl() string {
	return c.url
}

// GetInitialSequence returns the initial sequence and account number
func (c *cosmosRestClient) GetInitialSequence(ctx context.Context, address string) (uint64, uint64, error) {
	accountRes := &AccountResult{}
	url := fmt.Sprintf("%s/cosmos/auth/v1beta1/accounts/%s", c.GetUrl(), address)

	_, err := httputil.Get(ctx, url, nil, &accountRes)
	if err != nil {
		return 0, 0, err
	}

	var sequence, accountNumber string
	if accountRes.Account.BaseAccount != nil {
		// Injective format
		sequence = accountRes.Account.BaseAccount.Sequence
		accountNumber = accountRes.Account.BaseAccount.AccountNumber
	} else {
		// Standard Cosmos format
		sequence = accountRes.Account.Sequence
		accountNumber = accountRes.Account.AccountNumber
	}

	seqint, err := strconv.ParseUint(sequence, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	accnum, err := strconv.ParseUint(accountNumber, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return seqint, accnum, nil
}

// GetAllBalances returns all balances for an address
func (c *cosmosRestClient) GetAllBalances(ctx context.Context, address string) (BalancesResponse, error) {
	url := fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", c.GetUrl(), address)

	var balancesResp BalancesResponse
	_, err := httputil.Get(ctx, url, nil, &balancesResp)
	if err != nil {
		return BalancesResponse{}, fmt.Errorf("failed to get balances: %w", err)
	}

	return balancesResp, nil
}

// SimulateResponseGasInfo is a minimal struct to unmarshal only the gas_info
type SimulateResponseGasInfo struct {
	GasInfo struct {
		GasUsed uint64 `json:"gas_used,string"`
	} `json:"gas_info"`
}

// SimulateGasUsed simulates a transaction to estimate gas usage
func (c *cosmosRestClient) SimulateGasUsed(ctx context.Context, simulateReq *tx.SimulateRequest) (uint64, error) {
	url := fmt.Sprintf("%s/cosmos/tx/v1beta1/simulate", c.GetUrl())

	reqBody, err := json.Marshal(simulateReq)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal simulate request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return 0, fmt.Errorf("failed to create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("simulate request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var gasInfo SimulateResponseGasInfo
	err = json.Unmarshal(body, &gasInfo)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal gas_used from simulate response: %v", err)
	}

	return gasInfo.GasInfo.GasUsed, nil
}

// validateUrl checks if a URL is valid
func validateUrl(urlStr string) error {
	_, err := url.Parse(urlStr)
	return err
}

var _ CosmosRESTClient = &cosmosRestClient{}

package broadcastcosmos

// Coin is the sdk.Coin type that is used in the Cosmos SDK
type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type BalancesResponse struct {
	Balances []Coin `json:"balances"`
}

type BaseAccountInfo struct {
	Sequence      string `json:"sequence"`
	AccountNumber string `json:"account_number"`
}

type AccountResult struct {
	Account struct {
		Type string `json:"@type"`
		// For Injective EthAccount format
		BaseAccount *BaseAccountInfo `json:"base_account,omitempty"`
		// For standard Cosmos format
		Sequence      string `json:"sequence,omitempty"`
		AccountNumber string `json:"account_number,omitempty"`
	} `json:"account"`
}

type BaseFeeResult struct {
	BaseFee string `json:"base_fee"`
}

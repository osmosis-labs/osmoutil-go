package broadcastcosmos

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	osmoutilstx "github.com/osmosis-labs/osmoutil-go/tx"

	broadcasttypes "github.com/osmosis-labs/osmoutil-go/tx/broadcast/types"
)

type CosmosSigner interface {
	broadcasttypes.Signer

	// Address returns the address of the payer
	Address() sdk.AccAddress

	// GetNonceTracker returns the nonce tracker
	GetNonceTracker() osmoutilstx.NonceTrackerI

	// GetPayer returns the private key of the payer
	GetPayer() cryptotypes.PrivKey

	// GetPubKey returns the public key of the payer
	GetPubKey() cryptotypes.PubKey

	// GetBech32Prefix returns the bech32 address prefix of the signer
	GetBech32Prefix() string

	// GetNativeChainID returns the native ID of the chain. For example, "osmosis-1"
	GetNativeChainID() string

	// GetFeeDenom returns the fee denom of the chain this signer is on. For example, "uosmo".
	GetFeeDenom() string

	// SignTransaction signs a transaction
	SignTransaction(ctx context.Context, txBuilder client.TxBuilder, txConfig client.TxConfig, accnum, sequence uint64) error

	// SetNonceTracker sets the nonce tracker for the signer. Unset in constructor.
	SetNonceTracker(nonceTracker osmoutilstx.NonceTrackerI)
}

type cosmosSigner struct {
	nonceTracker  osmoutilstx.NonceTrackerI
	payer         cryptotypes.PrivKey
	bech32Prefix  string
	nativeChainID string
	feeDenom      string
}

// InitializeCosmosSigner initializes and configures a Cosmos signer for transaction broadcasting.
// It creates a signer with the provided funder private key, sets up the nonce tracker,
// and connects it to the appropriate endpoint manager.
func InitializeCosmosSigner(ctx context.Context, privateKeyHex string, clientConfig broadcasttypes.CosmosClientConfig, restClient CosmosRESTClient) (CosmosSigner, error) {
	// Create the signer
	signer, err := NewCosmosSigner(
		privateKeyHex,
		clientConfig.Bech32Prefix,
		clientConfig.NativeChainID,
		clientConfig.FeeTokenDenom,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cosmos signer for %s: %w", clientConfig.Name, err)
	}

	isCustomForceRefetchInterval := clientConfig.ForceRefetchInterval != 0
	isCustomRefetchTimeout := clientConfig.RefetchTimeout != 0

	if isCustomForceRefetchInterval != isCustomRefetchTimeout {
		return nil, fmt.Errorf("force refetch interval and refetch timeout must be set together. Either both custom or none of them.")
	}

	// Initialize nonce tracker
	nonceTracker := NewCosmosNonceTracker(signer.GetAddressString(), restClient)

	if isCustomForceRefetchInterval && isCustomRefetchTimeout {
		// Override the default force refetch interval and refetch timeout
		osmoutilstx.WithCustomIntervals(clientConfig.ForceRefetchInterval, clientConfig.RefetchTimeout)(nonceTracker)
	}

	// Force refetches the nonce
	if _, err := nonceTracker.ForceRefetch(ctx); err != nil {
		return nil, fmt.Errorf("failed to force refetch nonce for %s: %w", clientConfig.Name, err)
	}

	// Set the nonce tracker
	signer.SetNonceTracker(nonceTracker)

	return signer, nil
}

// NewCosmosSigner creates a new Cosmos signer
func NewCosmosSigner(payerPKHex string, bech32Prefix, nativeChainID, feeDenom string) (CosmosSigner, error) {
	// Decode the private key from hex to bytes
	privKeyPayerBytes, err := hex.DecodeString(payerPKHex)
	if err != nil {
		return nil, err
	}

	// Create a secp256k1 private key object
	var privKeyPayer cryptotypes.PrivKey = &secp256k1.PrivKey{Key: privKeyPayerBytes}

	return &cosmosSigner{

		// Note: must be set using WithNonceTracker()
		nonceTracker:  nil,
		payer:         privKeyPayer,
		bech32Prefix:  bech32Prefix,
		nativeChainID: nativeChainID,
		feeDenom:      feeDenom,
	}, nil
}

// SetNonceTracker sets the nonce tracker for the signer. Unset in constructor.
func (s *cosmosSigner) SetNonceTracker(nonceTracker osmoutilstx.NonceTrackerI) {
	s.nonceTracker = nonceTracker
}

// Address implements the CosmosSigner
func (s *cosmosSigner) Address() sdk.AccAddress {
	pubKey := s.payer.PubKey()

	// Convert the public key to an AccAddress
	return sdk.AccAddress(pubKey.Address())
}

// PubKey implements the CosmosSigner
func (s *cosmosSigner) PubKey() cryptotypes.PubKey {
	return s.payer.PubKey()
}

// GetNonceTracker implements the CosmosSigner
func (s *cosmosSigner) GetNonceTracker() osmoutilstx.NonceTrackerI {
	return s.nonceTracker
}

// GetPayer implements the CosmosSigner
func (s *cosmosSigner) GetPayer() cryptotypes.PrivKey {
	return s.payer
}

// GetAddressString implements the CosmosSigner
func (s *cosmosSigner) GetAddressString() string {
	fromAddr := sdk.AccAddress(s.payer.PubKey().Address())
	return sdk.MustBech32ifyAddressBytes(s.bech32Prefix, fromAddr)
}

// GetPubKey implements the CosmosSigner
func (s *cosmosSigner) GetPubKey() cryptotypes.PubKey {
	return s.PubKey()
}

// GetBech32Prefix implements the CosmosSigner
func (s *cosmosSigner) GetBech32Prefix() string {
	return s.bech32Prefix
}

// GetNativeChainID implements the CosmosSigner
func (s *cosmosSigner) GetNativeChainID() string {
	return s.nativeChainID
}

// GetFeeDenom implements the CosmosSigner
func (s *cosmosSigner) GetFeeDenom() string {
	return s.feeDenom
}

// signTransaction signs a transaction using the chain service's private key.
// It creates a SignerData object with the chain's ID and account details,
// then signs the transaction using SIGN_MODE_DIRECT signing mode.
func (s *cosmosSigner) SignTransaction(ctx context.Context, txBuilder client.TxBuilder, txConfig client.TxConfig, accnum, sequence uint64) error {
	signerData := authsigning.SignerData{
		ChainID:       s.nativeChainID,
		AccountNumber: accnum,
		Sequence:      sequence,
	}

	sigV2, err := tx.SignWithPrivKey(
		ctx, signing.SignMode_SIGN_MODE_DIRECT, signerData,
		txBuilder, s.payer, txConfig, sequence)
	if err != nil {
		return fmt.Errorf("couldn't sign: %v", err)
	}

	return txBuilder.SetSignatures(sigV2)
}

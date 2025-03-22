package broadcasttypes

// Signer is the interface for a signer.
type Signer interface {
	// GetAddressString returns the address of the signer as a string in the native format of the chain.
	GetAddressString() string
}

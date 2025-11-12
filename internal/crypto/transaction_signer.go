package crypto

import "github.com/decred/dcrd/dcrec/secp256k1/v4"

// TransactionSigner defines the interface for signing and verifying transactions.
type TransactionSigner interface {
	// SignTransaction signs a 32-byte transaction hash with a private key
	// and returns the resulting signature.
	SignTransaction(txHash []byte, privateKey *secp256k1.PrivateKey) []byte

	// VerifySignature verifies that a signature for a transaction hash was created
	// by the private key corresponding to the given public key.
	VerifySignature(txHash []byte, signature []byte, publicKey *secp256k1.PublicKey) bool
}

package crypto

import (
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// TransactionSignerImpl implements the TransactionSigner interface using ECDSA.
type TransactionSignerImpl struct{}

// NewTransactionSigner creates a new instance of TransactionSignerImpl.
func NewTransactionSigner() TransactionSigner {
	return &TransactionSignerImpl{}
}

// SignTransaction signs a 32-byte transaction hash with a private key.
func (ts *TransactionSignerImpl) SignTransaction(txHash []byte, privateKey *secp256k1.PrivateKey) []byte {
	// ecdsa.Sign creates a new digital signature.
	signature := ecdsa.Sign(privateKey, txHash)
	return signature.Serialize()
}

// VerifySignature verifies a signature against a transaction hash and public key.
func (ts *TransactionSignerImpl) VerifySignature(txHash []byte, signatureBytes []byte, publicKey *secp256k1.PublicKey) bool {
	signature, err := ecdsa.ParseDERSignature(signatureBytes)
	if err != nil {
		return false
	}

	// Verify the signature.
	return signature.Verify(txHash, publicKey)
}

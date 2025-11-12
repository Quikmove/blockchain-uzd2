package crypto

import (
	"crypto/sha256"
	"strings"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func GeneratePrivateKeyFromMnemonic(mnemonic string) [32]byte {
	cleanedMnemonic := strings.TrimSpace(strings.ToLower(mnemonic))

	// Hash the cleaned mnemonic to produce a 32-byte private key
	privateKey := sha256.Sum256([]byte(cleanedMnemonic))

	return privateKey
}

func DerivePublicKeyFromPrivateKey(privateKey [32]byte) []byte {
	privKey := secp256k1.PrivKeyFromBytes(privateKey[:])
	return privKey.PubKey().SerializeCompressed()
}

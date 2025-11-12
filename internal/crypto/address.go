package crypto

import (
	"crypto/sha256"

	"golang.org/x/crypto/ripemd160"
)

// GenerateAddress creates a 20-byte address from a public key using the
func GenerateAddress(publicKey []byte) [20]byte {

	sha256Hasher := sha256.New()
	sha256Hasher.Write(publicKey)
	sha256Hash := sha256Hasher.Sum(nil)

	ripemd160Hasher := ripemd160.New()
	ripemd160Hasher.Write(sha256Hash)
	addressBytes := ripemd160Hasher.Sum(nil)

	var address [20]byte
	copy(address[:], addressBytes)

	return address
}

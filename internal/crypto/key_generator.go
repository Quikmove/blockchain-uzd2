package crypto

import (
	"crypto/rand"
	"io"
)

// KeyGenerator defines the interface for generating cryptographic key pairs.
type KeyGenerator interface {
	GenerateKeyPair() (publicKey []byte, privateKey []byte, err error)
}

// SimpleKeyGenerator is a basic implementation of KeyGenerator.
type SimpleKeyGenerator struct{}

// NewKeyGenerator creates a new instance of SimpleKeyGenerator.
func NewKeyGenerator() KeyGenerator {
	return &SimpleKeyGenerator{}
}

// GenerateKeyPair generates a new public/private key pair.
func (kg *SimpleKeyGenerator) GenerateKeyPair() ([]byte, []byte, error) {
	pubKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, pubKey); err != nil {
		return nil, nil, err
	}

	privKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, privKey); err != nil {
		return nil, nil, err
	}

	return pubKey, privKey, nil
}

package crypto

import (
	"crypto/sha256"
)

// SHA256Hasher implements the Hasher interface using SHA256
type SHA256Hasher struct{}

// NewSHA256Hasher creates a new SHA256 hasher
func NewSHA256Hasher() *SHA256Hasher {
	return &SHA256Hasher{}
}

// Hash computes the SHA256 hash of the input data
func (h *SHA256Hasher) Hash(data []byte) [32]byte {
	hash := sha256.Sum256(data)
	return hash
}

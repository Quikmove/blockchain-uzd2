package blockchain

import "crypto/sha256"

type SHA256Hasher struct{}

func NewSHA256Hasher() *SHA256Hasher {
	return &SHA256Hasher{}
}

func (h *SHA256Hasher) Hash(data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	return hash[:], nil
}

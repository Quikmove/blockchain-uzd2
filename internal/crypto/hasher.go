package crypto

// Hasher defines the interface for hash functions
type Hasher interface {
	Hash(data []byte) [32]byte
}

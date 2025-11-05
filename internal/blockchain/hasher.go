package blockchain

// Hasher defines the interface for a hashing algorithm.
type Hasher interface {
	Hash(data []byte) ([]byte, error)
}

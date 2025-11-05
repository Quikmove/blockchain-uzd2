package blockchain

type Hasher interface {
	Hash(data []byte) ([]byte, error)
}

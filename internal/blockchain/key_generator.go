package blockchain

type KeyGenerator interface {
	GenerateKeyPair() (publicKey []byte, privateKey []byte, err error)
}

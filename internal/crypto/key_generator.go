package crypto

type KeyGenerator interface {
	GenerateKeyPair(mnemonic string) (privateKey [32]byte, publicKey [33]byte, err error)
}

type SimpleKeyGenerator struct{}

func NewKeyGenerator() *SimpleKeyGenerator {
	return &SimpleKeyGenerator{}
}

var _ KeyGenerator = (*SimpleKeyGenerator)(nil)

func (kg *SimpleKeyGenerator) GenerateKeyPair(mnemonic string) ([32]byte, [33]byte, error) {
	privateKey := GeneratePrivateKeyFromMnemonic(mnemonic)
	publicKeySlice := DerivePublicKeyFromPrivateKey(privateKey)

	var publicKeyArray [33]byte
	copy(publicKeyArray[:], publicKeySlice)

	return privateKey, publicKeyArray, nil
}

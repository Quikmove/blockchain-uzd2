package crypto

type TransactionSigner interface {
	SignTransaction(txData []byte, privateKey []byte) ([32]byte, error)
}

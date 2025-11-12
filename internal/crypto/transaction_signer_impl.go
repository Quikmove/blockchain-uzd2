package crypto

type TransactionSignerImpl struct {
	hasher Hasher
}

func NewTransactionSigner(hasher Hasher) *TransactionSignerImpl {
	return &TransactionSignerImpl{hasher: hasher}
}

func (ts *TransactionSignerImpl) SignTransaction(txData []byte, privateKey []byte) [32]byte {
	signature := ts.hasher.Hash(append(txData, privateKey...))
	return signature
}

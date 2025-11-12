package domain

import "errors"

var (
	// Block errors
	ErrInvalidBlock         = errors.New("invalid block")
	ErrInvalidPrevHash      = errors.New("previous hash mismatch")
	ErrInvalidDifficulty    = errors.New("hash does not meet difficulty requirements")
	ErrInvalidMerkleRoot    = errors.New("merkle root mismatch")
	ErrBlockNotFound        = errors.New("block not found")
	ErrBlockIndexOutOfRange = errors.New("block index out of range")
	ErrEmptyBlockchain      = errors.New("blockchain is empty")

	// Transaction errors
	ErrInvalidTransaction = errors.New("invalid transaction")
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrUTXONotFound       = errors.New("utxo not found")
	ErrDoubleSpend        = errors.New("double spend detected")
	ErrInvalidSignature   = errors.New("invalid signature")
	ErrEmptyTransaction   = errors.New("transaction has no outputs")

	// User errors
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidPublicKey = errors.New("invalid public key")

	// Hash errors
	ErrInvalidHashLength = errors.New("invalid hash length")

	// Mining errors
	ErrMiningCanceled = errors.New("mining operation canceled")
	ErrNoValidNonce   = errors.New("no valid nonce found")
)

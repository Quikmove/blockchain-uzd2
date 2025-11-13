package blockchain

import (
	"errors"
	"time"

	c "github.com/Quikmove/blockchain-uzd2/internal/crypto"
	d "github.com/Quikmove/blockchain-uzd2/internal/domain"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func (bch *Blockchain) IsBlockValid(newBlock d.Block) bool {
	bch.chainMutex.RLock()
	height := len(bch.blocks)
	bch.chainMutex.RUnlock()
	if height == 0 {
		return true
	}
	oldBlock, err := bch.GetLatestBlock()
	if err != nil {
		panic(err)
	}

	oldBlockHash := bch.CalculateHash(oldBlock)
	header := newBlock.Header
	if oldBlockHash != header.PrevHash {
		return false
	}

	diff := header.Difficulty
	hash := bch.CalculateHash(newBlock)

	return IsHashValid(hash, diff)
}

func (bch *Blockchain) ValidateBlock(b d.Block) error {
	bch.chainMutex.RLock()
	height := len(bch.blocks)
	bch.chainMutex.RUnlock()

	isGenesis := height == 0

	// Validate block has transactions
	body := b.Body
	txs := body.Transactions
	if len(txs) == 0 {
		return d.ErrInvalidBlock
	}

	// Validate merkle root
	computedMerkleRoot := MerkleRootHash(body, bch.hasher)
	if computedMerkleRoot != b.Header.MerkleRoot {
		return d.ErrInvalidMerkleRoot
	}

	// Validate block hash meets difficulty (for non-genesis blocks)
	if !isGenesis {
		hash := bch.CalculateHash(b)
		if !IsHashValid(hash, b.Header.Difficulty) {
			return d.ErrInvalidDifficulty
		}
	}

	currentTime := uint32(time.Now().Unix())
	maxFutureTime := currentTime + 7200
	minPastTime := currentTime - 7200
	if b.Header.Timestamp > maxFutureTime {
		return errors.New("block timestamp too far in future")
	}
	if !isGenesis && b.Header.Timestamp < minPastTime {
		return errors.New("block timestamp too far in past")
	}

	for i, tx := range txs {
		expectedTxID := bch.hasher.Hash(tx.SerializeWithoutSignatures())
		if tx.TxID != expectedTxID {
			return d.ErrInvalidTransaction
		}

		isCoinbase := len(tx.Inputs) == 0
		if isGenesis {
			if !isCoinbase {
				return d.ErrInvalidTransaction
			}
			continue
		}
		if isCoinbase {
			if i != 0 {
				return d.ErrInvalidTransaction
			}
			continue
		}

		if len(tx.Inputs) == 0 {
			return d.ErrInvalidTransaction
		}
	}

	return nil
}

func (bch *Blockchain) ValidateBlockTransactions(b d.Block, users []d.User) error {
	bch.chainMutex.RLock()
	height := len(bch.blocks)
	bch.chainMutex.RUnlock()

	isGenesis := height == 0

	body := b.Body
	txs := body.Transactions
	if len(txs) == 0 {
		return d.ErrInvalidBlock
	}

	spentInBlock := make(map[d.Outpoint]bool)

	addressToPublicKey := make(map[d.PublicAddress]d.PublicKey)
	for _, user := range users {
		addressToPublicKey[user.PublicAddress] = user.PublicKey
	}

	for i, tx := range txs {
		isCoinbase := len(tx.Inputs) == 0

		if isGenesis && !isCoinbase {
			return d.ErrInvalidTransaction
		}

		if isCoinbase {
			if i != 0 {
				return d.ErrInvalidTransaction
			}
			if len(tx.Outputs) == 0 {
				return d.ErrInvalidTransaction
			}
			var coinbaseTotal uint32
			for _, output := range tx.Outputs {
				if output.Value == 0 {
					return d.ErrInvalidTransaction
				}
				if coinbaseTotal > ^uint32(0)-output.Value {
					return errors.New("coinbase tx total reward overflow")
				}
				coinbaseTotal += output.Value
			}
			continue
		}

		if len(tx.Inputs) == 0 {
			return d.ErrInvalidTransaction
		}

		if len(tx.Outputs) == 0 {
			return d.ErrEmptyTransaction
		}

		var inputSum uint32
		for _, input := range tx.Inputs {
			if spentInBlock[input.Prev] {
				return d.ErrDoubleSpend
			}

			utxo, exists := bch.utxoTracker.GetUTXO(input.Prev)
			if !exists {
				return d.ErrUTXONotFound
			}

			if inputSum > ^uint32(0)-utxo.Value {
				return d.ErrNoValidNonce
			}
			inputSum += utxo.Value

			if !isGenesis {
				if len(input.Sig) == 0 {
					return errors.New("missing signature for non-genesis transaction")
				}

				publicKey, hasKey := addressToPublicKey[utxo.To]
				if !hasKey {
					for _, user := range users {
						if user.PublicAddress == utxo.To {
							publicKey = user.PublicKey
							hasKey = true
							break
						}
					}
				}

				if !hasKey {
					return d.ErrInvalidPublicKey
				}

				expectedAddress := c.GenerateAddress(publicKey[:])
				if utxo.To != expectedAddress {
					return d.ErrInvalidPublicKey
				}

				hashToVerify := SignatureHash(tx, utxo.Value, utxo.To[:], bch.hasher)

				publicKeyObj, err := secp256k1.ParsePubKey(publicKey[:])
				if err != nil {
					return d.ErrInvalidPublicKey
				}

				if !bch.txSigner.VerifySignature(hashToVerify[:], input.Sig, publicKeyObj) {
					return d.ErrInvalidSignature
				}
			}

			spentInBlock[input.Prev] = true
		}

		var outputSum uint32
		for _, output := range tx.Outputs {
			if output.Value == 0 {
				return errors.New("zero-value output not allowed")
			}

			if outputSum > ^uint32(0)-output.Value {
				return errors.New("output sum overflow")
			}
			outputSum += output.Value
		}

		if inputSum < outputSum {
			return d.ErrInsufficientFunds
		}
	}

	return nil
}

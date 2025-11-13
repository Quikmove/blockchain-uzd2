package blockchain

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/Quikmove/blockchain-uzd2/internal/config"
	c "github.com/Quikmove/blockchain-uzd2/internal/crypto"
	d "github.com/Quikmove/blockchain-uzd2/internal/domain"
)

type Blockchain struct {
	blocks       []d.Block
	chainMutex   *sync.RWMutex
	txGenMutex   *sync.Mutex
	utxoTracker  *UTXOTracker
	hasher       c.Hasher
	txSigner     c.TransactionSigner
	userRegistry map[d.PublicAddress]d.PublicKey
	userMutex    *sync.RWMutex
}

func NewBlockchain(hasher c.Hasher, signer c.TransactionSigner) *Blockchain {
	return &Blockchain{
		blocks:       []d.Block{},
		chainMutex:   &sync.RWMutex{},
		txGenMutex:   &sync.Mutex{},
		utxoTracker:  NewUTXOTracker(),
		hasher:       hasher,
		txSigner:     signer,
		userRegistry: make(map[d.PublicAddress]d.PublicKey),
		userMutex:    &sync.RWMutex{},
	}
}

func (bch *Blockchain) GetBlock(index int) (d.Block, error) {
	bch.chainMutex.RLock()
	defer bch.chainMutex.RUnlock()
	if index < 0 || index >= len(bch.blocks) {
		return d.Block{}, d.ErrBlockIndexOutOfRange
	}
	return bch.blocks[index], nil
}
func (bch *Blockchain) GetLatestBlock() (d.Block, error) {
	bch.chainMutex.RLock()
	defer bch.chainMutex.RUnlock()
	if len(bch.blocks) == 0 {
		return d.Block{}, d.ErrEmptyBlockchain
	}
	return bch.blocks[len(bch.blocks)-1], nil
}
func (bch *Blockchain) AddBlock(b d.Block) error {
	if err := bch.ValidateBlock(b); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}

	users := bch.getUsersFromRegistry()
	if err := bch.ValidateBlockTransactions(b, users); err != nil {
		return fmt.Errorf("block transaction validation failed: %w", err)
	}

	bch.chainMutex.Lock()
	defer bch.chainMutex.Unlock()

	height := len(bch.blocks)

	if height != 0 {
		tip := bch.blocks[height-1]
		tipHash := bch.CalculateHash(tip)
		header := b.Header
		if tipHash != header.PrevHash {
			return d.ErrInvalidPrevHash
		}
		hash := bch.CalculateHash(b)
		if !IsHashValid(hash, header.Difficulty) {
			return d.ErrInvalidDifficulty
		}
	}

	bch.blocks = append(bch.blocks, b)

	bch.utxoTracker.ScanBlock(b, bch.hasher)

	return nil
}

func (bch *Blockchain) Blocks() []d.Block {
	bch.chainMutex.RLock()
	defer bch.chainMutex.RUnlock()
	var blocksCopy = make([]d.Block, len(bch.blocks))

	for i, b := range bch.blocks {
		var bodyCopy d.Body
		body := b.Body
		txs := body.Transactions
		if len(txs) > 0 {
			bodyCopy.Transactions = (make([]d.Transaction, len(txs)))
			for j, tx := range txs {
				var inputs []d.TxInput
				if len(tx.Inputs) > 0 {
					inputs = make([]d.TxInput, len(tx.Inputs))
					for k, in := range tx.Inputs {
						sigCopy := make([]byte, len(in.Sig))
						copy(sigCopy, in.Sig)
						inputs[k] = d.TxInput{
							Prev: d.Outpoint{
								TxID:  in.Prev.TxID,
								Index: in.Prev.Index,
							},
							Sig: sigCopy,
						}
					}
				}

				var outputs []d.TxOutput
				if len(tx.Outputs) > 0 {
					outputs = make([]d.TxOutput, len(tx.Outputs))
					copy(outputs, tx.Outputs)
				}
				bodyCopy.Transactions[j] = d.Transaction{
					TxID:    tx.TxID,
					Inputs:  inputs,
					Outputs: outputs,
				}
			}
		} else {
			bodyCopy.Transactions = (nil)
		}
		blocksCopy[i] = d.Block{
			Header: b.Header,
			Body:   bodyCopy,
		}
	}
	return blocksCopy
}
func (bch *Blockchain) Len() int {
	bch.chainMutex.RLock()
	defer bch.chainMutex.RUnlock()
	return len(bch.blocks)
}
func InitBlockchainWithFunds(low, high uint32, users []d.User, cfg *config.Config, hasher c.Hasher, txSigner c.TransactionSigner) *Blockchain {
	fundTransactions, err := GenerateFundTransactionsForUsers(users, low, high, hasher)
	if err != nil {
		panic(err)
	}
	genesisBlock, err := CreateGenesisBlock(context.Background(), fundTransactions, cfg, hasher)
	if err != nil {
		panic(err)
	}
	blockchain := NewBlockchain(hasher, txSigner)
	blockchain.RegisterUsers(users)
	blockchain.blocks = append(blockchain.blocks, genesisBlock)

	blockchain.utxoTracker.ScanBlock(genesisBlock, hasher)

	return blockchain
}

func MerkleRootHash(b d.Body, hasher c.Hasher) d.Hash32 {
	return merkleRootHash(b.Transactions, hasher)
}

type Transactions []d.Transaction

func HashBytes(bytes []byte, hasher c.Hasher) d.Hash32 {
	h := hasher.Hash(bytes)

	return h
}

func HashString(str string, hasher c.Hasher) (d.Hash32, error) {
	h := HashBytes([]byte(str), hasher)

	return h, nil
}

func (bch *Blockchain) CalculateHash(block d.Block) d.Hash32 {
	if block.Header.PrevHash.IsZero() && block.Header.MerkleRoot.IsZero() {
		return d.Hash32{}
	}
	hash := bch.hasher.Hash(block.Header.Serialize())
	return hash
}
func IsHashValid(hash d.Hash32, diff uint32) bool {
	if diff == 0 {
		return true
	}

	bits := diff * 4
	hashLen := len(hash)
	maxBits := uint32(hashLen) * 8
	if bits > maxBits {
		return false
	}
	fullBytes := bits / 8
	remBits := bits % 8

	if fullBytes > 0 {
		var zero [32]byte
		if !bytes.Equal(hash[:fullBytes], zero[:fullBytes]) {
			return false
		}
	}

	if remBits == 0 {
		return true
	}

	byteIdx := int(fullBytes)
	if byteIdx >= hashLen {
		return false
	}
	mask := byte(0xFF << (8 - remBits))
	return (hash[byteIdx] & mask) == 0
}

func (bch *Blockchain) GenerateRandomTransactions(users []d.User, low, high, n int) (Transactions, error) {
	bch.txGenMutex.Lock()
	defer bch.txGenMutex.Unlock()

	if high < low || low < 0 {
		return nil, errors.New("invalid amount range")
	}
	if len(users) < 2 {
		return nil, errors.New("not enough users to generate transactions")
	}

	var generatedTxs Transactions
	userAmount := len(users)
	usedOutpoints := make(map[d.Outpoint]bool)

	maxAttempts := n * 10
	attempts := 0

	for len(generatedTxs) < n && attempts < maxAttempts {
		attempts++

		senderIndex := rand.Intn(userAmount)
		recipientIndex := rand.Intn(userAmount)
		for senderIndex == recipientIndex {
			recipientIndex = rand.Intn(userAmount)
		}
		sender := users[senderIndex]
		recipient := users[recipientIndex]

		utxos := bch.utxoTracker.GetUTXOsForAddress(sender.PublicAddress)

		if len(utxos) == 0 {
			continue
		}

		amount := uint32(low + rand.Intn(high-low+1))
		if amount == 0 {
			continue
		}

		var inputs []d.TxInput
		var selectedUTXOs []d.UTXO
		var totalInput uint32

		for _, utxo := range utxos {
			if totalInput >= amount {
				break
			}
			if usedOutpoints[utxo.Outpoint] {
				continue
			}
			if totalInput > ^uint32(0)-utxo.Value {
				continue
			}
			inputs = append(inputs, d.TxInput{Prev: utxo.Outpoint})
			selectedUTXOs = append(selectedUTXOs, utxo)
			totalInput += utxo.Value
		}

		if len(inputs) == 0 {
			continue
		}

		if totalInput < amount {
			continue
		}

		var outputs []d.TxOutput
		outputs = append(outputs, d.TxOutput{Value: amount, To: recipient.PublicAddress})

		if totalInput > amount {
			change := totalInput - amount
			outputs = append(outputs, d.TxOutput{Value: change, To: sender.PublicAddress})
		}

		tx := d.Transaction{
			Inputs:  inputs,
			Outputs: outputs,
		}

		txID := bch.hasher.Hash(tx.SerializeWithoutSignatures())
		tx.TxID = txID

		for j := range tx.Inputs {
			hashToSign := SignatureHash(tx, selectedUTXOs[j].Value, selectedUTXOs[j].To[:], bch.hasher)
			sig := bch.txSigner.SignTransaction(hashToSign[:], sender.GetPrivateKeyObject())
			tx.Inputs[j].Sig = sig[:]
		}

		generatedTxs = append(generatedTxs, tx)
		for _, utxo := range selectedUTXOs {
			usedOutpoints[utxo.Outpoint] = true
		}
	}

	if len(generatedTxs) == 0 {
		return nil, d.ErrInsufficientFunds
	}

	return generatedTxs, nil
}
func (bch *Blockchain) GenerateBlock(ctx context.Context, body d.Body, version uint32, difficulty uint32) (d.Block, error) {
	latestBlock, err := bch.GetLatestBlock()
	if err != nil {
		return d.Block{}, err
	}
	var newHeader d.Header
	t := time.Now()

	newHeader.Version = version
	newHeader.Timestamp = uint32(t.Unix())
	newHeader.PrevHash = bch.CalculateHash(latestBlock)
	newHeader.MerkleRoot = MerkleRootHash(body, bch.hasher)
	newHeader.Difficulty = difficulty

	nonce, _, err := FindValidNonce(ctx, &newHeader, bch.hasher)
	if err != nil {
		return d.Block{}, err
	}
	newHeader.Nonce = nonce

	newBlock := d.Block{
		Header: newHeader,
		Body:   body,
	}

	return newBlock, nil
}
func (bch *Blockchain) GetBlockByIndex(index int) (d.Block, error) {
	bch.chainMutex.RLock()
	defer bch.chainMutex.RUnlock()
	if index < 0 || index >= len(bch.blocks) {
		return d.Block{}, d.ErrBlockIndexOutOfRange
	}
	return bch.blocks[index], nil
}
func (bch *Blockchain) Print(w io.Writer) error {
	blocks := bch.Blocks()
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(blocks)
}
func (bch *Blockchain) GetUserBalance(address d.PublicAddress) uint32 {
	bch.chainMutex.RLock()
	defer bch.chainMutex.RUnlock()
	balance := bch.utxoTracker.GetBalance(address)
	return balance
}
func (bch *Blockchain) GetUTXOsForAddress(address d.PublicAddress) []d.UTXO {
	return bch.utxoTracker.GetUTXOsForAddress(address)
}

func (bch *Blockchain) RegisterUsers(users []d.User) {
	bch.userMutex.Lock()
	defer bch.userMutex.Unlock()
	for _, user := range users {
		bch.userRegistry[user.PublicAddress] = user.PublicKey
	}
}

func (bch *Blockchain) getUsersFromRegistry() []d.User {
	bch.userMutex.RLock()
	defer bch.userMutex.RUnlock()

	users := make([]d.User, 0, len(bch.userRegistry))
	for address, publicKey := range bch.userRegistry {
		users = append(users, d.User{
			PublicAddress: address,
			PublicKey:     publicKey,
		})
	}
	return users
}

func (bch *Blockchain) String() string {
	blocks := bch.Blocks()
	b, err := json.MarshalIndent(blocks, "", "  ")
	if err != nil {
		return ""
	}
	return string(b)
}

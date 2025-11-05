package blockchain

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Quikmove/blockchain-uzd2/internal/config"
)

var userCount atomic.Uint32

type Blockchain struct {
	blocks      []Block
	ChainMutex  *sync.RWMutex
	utxoTracker *UTXOTracker
	hasher      Hasher
}

func NewBlockchain(hasher Hasher) *Blockchain {
	return &Blockchain{
		blocks:      []Block{},
		ChainMutex:  &sync.RWMutex{},
		utxoTracker: NewUTXOTracker(),
		hasher:      hasher,
	}
}

func (bc *Blockchain) GetBlock(index int) (Block, error) {
	bc.ChainMutex.RLock()
	defer bc.ChainMutex.RUnlock()
	if index < 0 || index >= len(bc.blocks) {
		return Block{}, errors.New("block index out of range")
	}
	return bc.blocks[index], nil
}
func (bc *Blockchain) GetLatestBlock() (Block, error) {
	bc.ChainMutex.RLock()
	defer bc.ChainMutex.RUnlock()
	if len(bc.blocks) == 0 {
		return Block{}, errors.New("blockchain is empty")
	}
	return bc.blocks[len(bc.blocks)-1], nil
}
func (bc *Blockchain) AddBlock(b Block) error {
	bc.ChainMutex.RLock()
	if !bc.IsBlockValid(b) {
		bc.ChainMutex.RUnlock()
		return errors.New("invalid block")
	}
	bc.ChainMutex.RUnlock()

	if err := bc.ValidateBlockTransactions(b); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}

	bc.ChainMutex.Lock()
	defer bc.ChainMutex.Unlock()

	bc.blocks = append(bc.blocks, b)

	bc.utxoTracker.ScanBlock(b, bc.hasher)

	return nil
}

func (bc *Blockchain) Blocks() []Block {
	bc.ChainMutex.RLock()
	defer bc.ChainMutex.RUnlock()
	var blocksCopy = make([]Block, len(bc.blocks))

	for i, b := range bc.blocks {
		var bodyCopy Body
		if len(b.Body.Transactions) > 0 {
			bodyCopy.Transactions = make([]Transaction, len(b.Body.Transactions))
			for j, tx := range b.Body.Transactions {
				var inputs []TxInput
				if len(tx.Inputs) > 0 {
					inputs = make([]TxInput, len(tx.Inputs))
					for k, in := range tx.Inputs {
						sigCopy := make([]byte, len(in.Sig))
						copy(sigCopy, in.Sig)
						inputs[k] = TxInput{
							Prev: Outpoint{
								TxID:  in.Prev.TxID,
								Index: in.Prev.Index,
							},
							Sig: sigCopy,
						}
					}
				}

				var outputs []TxOutput
				if len(tx.Outputs) > 0 {
					outputs = make([]TxOutput, len(tx.Outputs))
					copy(outputs, tx.Outputs)
				}
				bodyCopy.Transactions[j] = Transaction{
					TxID:    tx.TxID,
					Inputs:  inputs,
					Outputs: outputs,
				}
			}
		} else {
			bodyCopy.Transactions = nil
		}
		blocksCopy[i] = Block{
			Header: b.Header,
			Body:   bodyCopy,
		}
	}
	return blocksCopy
}
func (bc *Blockchain) Len() int {
	bc.ChainMutex.RLock()
	defer bc.ChainMutex.RUnlock()
	return len(bc.blocks)
}
func InitBlockchainWithFunds(low, high uint32, users []User, cfg *config.Config, hasher Hasher) *Blockchain {
	fundTransactions, err := GenerateFundTransactionsForUsers(users, low, high, hasher)
	if err != nil {
		panic(err)
	}
	genesisBlock, err := CreateGenesisBlock(context.Background(), fundTransactions, cfg, hasher)
	if err != nil {
		panic(err)
	}
	blockchain := NewBlockchain(hasher)
	blockchain.blocks = append(blockchain.blocks, genesisBlock)

	blockchain.utxoTracker.ScanBlock(genesisBlock, hasher)

	return blockchain
}

type Hash32 [32]byte

func (h Hash32) HexString() string {
	return hex.EncodeToString(h[:])
}
func (h Hash32) MarshalJSON() ([]byte, error) {
	s := "\"" + h.HexString() + "\""
	return []byte(s), nil
}

type Block struct {
	Header Header `json:"header"`
	Body   Body   `json:"body"`
}
type Header struct {
	Version    uint32 `json:"version"`
	Timestamp  uint32 `json:"timestamp"`
	PrevHash   Hash32 `json:"prev_hash"`
	MerkleRoot Hash32 `json:"merkle_root"`
	Difficulty uint32 `json:"difficulty"`
	Nonce      uint32 `json:"nonce"`
}
type Body struct {
	Transactions Transactions `json:"transactions"`
}

func (h Header) Serialize() []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.LittleEndian, h.Version)
	buf.Write(reverse32(h.PrevHash))
	buf.Write(reverse32(h.MerkleRoot))
	_ = binary.Write(&buf, binary.LittleEndian, h.Timestamp)
	_ = binary.Write(&buf, binary.LittleEndian, h.Difficulty)
	_ = binary.Write(&buf, binary.LittleEndian, h.Nonce)

	return buf.Bytes()
}
func (h Header) Hash(hasher Hasher) (Hash32, error) {
	b := h.Serialize()
	x, err := hasher.Hash(b)
	if err != nil {
		return Hash32{}, err
	}
	y, err := hasher.Hash(x)
	if err != nil {
		return Hash32{}, err
	}
	var hash Hash32
	copy(hash[:], y)
	return hash, nil
}
func reverse32(in Hash32) []byte {
	out := make([]byte, 32)
	for i := 0; i < 32; i++ {
		out[i] = in[31-i]
	}
	return out
}
func (b Body) MerkleRootHash(hasher Hasher) Hash32 {
	return merkleRootHash(b.Transactions, hasher)
}

type Transactions []Transaction

func HashBytes(bytes []byte, hasher Hasher) (Hash32, error) {
	h, err := hasher.Hash(bytes)
	if err != nil {
		return Hash32{}, err
	}
	var hash32 Hash32
	copy(hash32[:], h)
	return hash32, nil
}

func HashString(str string, hasher Hasher) (Hash32, error) {
	h, err := HashBytes([]byte(str), hasher)
	if err != nil {
		return Hash32{}, err
	}
	return h, nil
}

func (bc *Blockchain) CalculateHash(block Block) (Hash32, error) {
	hash, err := block.Header.Hash(bc.hasher)
	if err != nil {
		return Hash32{}, err
	}
	return hash, nil
}
func IsHashValid(hash Hash32, diff uint32) bool {
	for i := uint32(0); i < diff && i < 32; i++ {
		if hash[i] != 0 {
			return false
		}
	}
	return true
}
func (bc *Blockchain) IsBlockValid(newBlock Block) bool {
	bc.ChainMutex.RLock()
	height := len(bc.blocks)
	bc.ChainMutex.RUnlock()
	if height == 0 {
		return true
	}
	oldBlock, err := bc.GetLatestBlock()
	if err != nil {
		panic(err)
	}

	oldBlockHash, err := bc.CalculateHash(oldBlock)
	if err != nil {
		return false
	}
	if oldBlockHash != newBlock.Header.PrevHash {
		return false
	}

	diff := newBlock.Header.Difficulty
	hash, err := bc.CalculateHash(newBlock)
	if err != nil {
		return false
	}

	return IsHashValid(hash, diff)
}

func (bc *Blockchain) ValidateBlock(b Block) error {
	bc.ChainMutex.RLock()
	height := len(bc.blocks)
	bc.ChainMutex.RUnlock()

	isGenesis := height == 0

	if len(b.Body.Transactions) == 0 {
		return errors.New("block has no transactions")
	}

	for i, tx := range b.Body.Transactions {
		isCoinbase := len(tx.Inputs) == 0
		if isGenesis {
			if !isCoinbase {
				return fmt.Errorf("genesis tx %d must be coinbase-like", i)
			}
			continue
		}
		if isCoinbase {
			if i != 0 {
				return fmt.Errorf("coinbase tx only allowed as first tx (found at index %d)", i)
			}
			continue
		}

		if len(tx.Inputs) == 0 {
			return fmt.Errorf("non-coinbase tx has no inputs (index %d)", i)
		}

	}
	return nil
}

func (bc *Blockchain) ValidateBlockTransactions(b Block) error {
	bc.ChainMutex.RLock()
	height := len(bc.blocks)
	bc.ChainMutex.RUnlock()

	isGenesis := height == 0

	if len(b.Body.Transactions) == 0 {
		return errors.New("block has no transactions")
	}

	spentInBlock := make(map[Outpoint]bool)

	for i, tx := range b.Body.Transactions {
		isCoinbase := len(tx.Inputs) == 0

		if isGenesis && !isCoinbase {
			return fmt.Errorf("genesis tx %d must be coinbase-like", i)
		}

		if isCoinbase {
			if i != 0 {
				return fmt.Errorf("coinbase tx only allowed as first tx (found at index %d)", i)
			}
			continue
		}

		if len(tx.Inputs) == 0 {
			return fmt.Errorf("non-coinbase tx has no inputs (index %d)", i)
		}

		var inputSum uint32
		for inputIdx, input := range tx.Inputs {
			if spentInBlock[input.Prev] {
				return fmt.Errorf("tx %d input %d: double-spend detected within block", i, inputIdx)
			}

			utxo, exists := bc.utxoTracker.GetUTXO(input.Prev)
			if !exists {
				return fmt.Errorf("tx %d input %d: references non-existent UTXO %s:%d",
					i, inputIdx, input.Prev.TxID.HexString(), input.Prev.Index)
			}

			if inputSum+utxo.Value < inputSum {
				return fmt.Errorf("tx %d: input sum overflow", i)
			}
			inputSum += utxo.Value

			spentInBlock[input.Prev] = true
		}

		var outputSum uint32
		for outputIdx, output := range tx.Outputs {
			if outputSum+output.Value < outputSum {
				return fmt.Errorf("tx %d: output sum overflow", i)
			}
			outputSum += output.Value

			if output.Value == 0 {
				return fmt.Errorf("tx %d output %d: zero-value output not allowed", i, outputIdx)
			}
		}

		if inputSum < outputSum {
			return fmt.Errorf("tx %d: outputs (%d) exceed inputs (%d)", i, outputSum, inputSum)
		}
	}

	return nil
}
func (h Header) FindValidNonce(ctx context.Context, hasher Hasher) (uint32, Hash32, error) {
	if h.Difficulty == 0 {
		hash, err := h.Hash(hasher)
		if err != nil {
			return 0, Hash32{}, err
		}
		return h.Nonce, hash, nil
	}
	if h.MerkleRoot == (Hash32{}) {
		return 0, Hash32{}, errors.New("merkle root not set")
	}

	var nonce uint32

	for {
		select {
		case <-ctx.Done():
			return 0, Hash32{}, ctx.Err()
		default:
			h.Nonce = nonce
			hash, err := h.Hash(hasher)
			if err != nil {
				return 0, Hash32{}, err
			}
			if IsHashValid(hash, h.Difficulty) {
				return nonce, hash, nil
			}
			nonce++

			if nonce == 0 {
				return 0, Hash32{}, errors.New("nonce overflow")
			}
		}
	}
}
func (bc *Blockchain) GenerateRandomTransactions(users []User, low, high, n int) (Transactions, error) {
	if high < low {
		return nil, errors.New("invalid amount range")
	}
	if len(users) < 2 {
		return nil, errors.New("not enough users to generate transactions")
	}

	var generatedTxs Transactions
	userAmount := len(users)
	usedOutpoints := make(map[Outpoint]bool, 0)

	for i := 0; i < n; i++ {
		senderIndex := rand.Intn(userAmount)
		recipientIndex := rand.Intn(userAmount)
		for senderIndex == recipientIndex {
			recipientIndex = rand.Intn(userAmount)
		}
		sender := users[senderIndex]
		recipient := users[recipientIndex]

		utxos := bc.utxoTracker.GetUTXOsForAddress(sender.PublicKey)

		if len(utxos) == 0 {
			continue
		}

		amount := uint32(low + rand.Intn(high-low+1))

		var inputs []TxInput
		var selectedUTXOs []UTXO
		var totalInput uint32

		for _, utxo := range utxos {
			if totalInput >= amount {
				break
			}
			if usedOutpoints[utxo.Out] {
				continue
			}
			inputs = append(inputs, TxInput{Prev: utxo.Out})
			selectedUTXOs = append(selectedUTXOs, utxo)
			totalInput += utxo.Value
		}

		var outputs []TxOutput
		outputs = append(outputs, TxOutput{Value: amount, To: recipient.PublicKey})

		if totalInput > amount {
			change := totalInput - amount
			outputs = append(outputs, TxOutput{Value: change, To: sender.PublicKey})
		}

		tx := Transaction{Outputs: outputs}

		for j := range inputs {
			hashToSign, err := tx.SignatureHash(selectedUTXOs[j].Value, selectedUTXOs[j].To, bc.hasher)
			if err != nil {
				return nil, fmt.Errorf("failed to create signature hash: %w", err)
			}
			sig, err := sender.Sign(hashToSign)
			if err != nil {
				return nil, fmt.Errorf("failed to sign transaction input: %w", err)
			}
			inputs[j].Sig = sig
		}

		tx.Inputs = inputs
		txID, err := tx.Hash(bc.hasher)
		if err != nil {
			return nil, fmt.Errorf("failed to hash transaction: %w", err)
		}
		tx.TxID = txID

		generatedTxs = append(generatedTxs, tx)
		for _, utxo := range selectedUTXOs {
			usedOutpoints[utxo.Out] = true
		}
	}
	return generatedTxs, nil
}
func (bc *Blockchain) GenerateBlock(ctx context.Context, body Body, version uint32, difficulty uint32) (Block, error) {
	latestBlock, err := bc.GetLatestBlock()
	if err != nil {
		return Block{}, err
	}
	var newHeader Header
	t := time.Now()

	newHeader.Version = version
	prevHash, err := bc.CalculateHash(latestBlock)
	if err != nil {
		return Block{}, err
	}
	newHeader.PrevHash = prevHash
	newHeader.Timestamp = uint32(t.Unix())
	newHeader.MerkleRoot = body.MerkleRootHash(bc.hasher)
	newHeader.Difficulty = difficulty
	nonce, _, err := newHeader.FindValidNonce(ctx, bc.hasher)
	if err != nil {
		return Block{}, err
	}
	newHeader.Nonce = nonce

	newBlock := Block{
		Header: newHeader,
		Body:   body,
	}

	return newBlock, nil
}

package blockchain

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Quikmove/blockchain-uzd2/internal/config"
)

var userCount atomic.Uint32

type Blockchain struct {
	blocks     []Block
	ChainMutex *sync.RWMutex
	UTXOSet    map[Outpoint]UTXO
	UTXOMutex  *sync.RWMutex
}

func NewBlockchain() *Blockchain {
	return &Blockchain{
		blocks:     []Block{},
		ChainMutex: &sync.RWMutex{},
		UTXOSet:    make(map[Outpoint]UTXO),
		UTXOMutex:  &sync.RWMutex{},
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
	bc.ChainMutex.Lock()
	defer bc.ChainMutex.Unlock()
	if !bc.IsBlockValid(b) {
		return errors.New("invalid block")
	}
	bc.blocks = append(bc.blocks, b)
	return nil
}
func (bc *Blockchain) Blocks() []Block {
	bc.ChainMutex.RLock()
	defer bc.ChainMutex.RUnlock()
	var blocksCopy = make([]Block, len(bc.blocks))
	copy(blocksCopy, bc.blocks)
	return blocksCopy
}
func (bc *Blockchain) Len() int {
	bc.ChainMutex.RLock()
	defer bc.ChainMutex.RUnlock()
	return len(bc.blocks)
}
func InitBlockchainWithFunds(low, high uint32, users []User, cfg *config.Config) *Blockchain {
	fundTransactions, err := GenerateFundTransactionsForUsers(users, low, high)
	if err != nil {
		panic(err)
	}
	genesisBlock, err := CreateGenesisBlock(context.Background(), fundTransactions, cfg)
	if err != nil {
		panic(err)
	}
	blockchain := NewBlockchain()
	blockchain.blocks = append(blockchain.blocks, genesisBlock)
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
func (h Header) Hash() Hash32 {
	b := h.Serialize()
	x := sha256.Sum256(b)
	return sha256.Sum256(x[:])
}
func reverse32(in Hash32) []byte {
	out := make([]byte, 32)
	for i := 0; i < 32; i++ {
		out[i] = in[31-i]
	}
	return out
}
func (b Body) MerkleRootHash() Hash32 {
	return MerkleRootHash(b.Transactions)
}

type Transactions []Transaction

func HashBytes(bytes []byte) Hash32 {
	return sha256.Sum256(bytes)
}

func HashString(str string) Hash32 {
	return HashBytes([]byte(str))
}

func CalculateHash(block Block) Hash32 {
	// use 6 main header properties: prev block hash, timestamp, version, merkel root hash, nonce, difficulty target
	hash := block.Header.Hash()
	return hash
}
func IsHashValid(hash Hash32, diff uint32) bool {
	for i := uint32(0); i < diff && i < 32; i++ {
		if hash[i] != 0 {
			return false
		}
	}
	return true
}
func (b *Blockchain) IsBlockValid(newBlock Block) bool {
	oldBlock := b.blocks[len(b.blocks)-1]
	if CalculateHash(oldBlock) != newBlock.Header.PrevHash {
		return false
	}

	diff := newBlock.Header.Difficulty
	hash := CalculateHash(newBlock)

	if IsHashValid(hash, diff) {
		return true
	}

	return true
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

		if !isCoinbase && len(tx.Inputs) == 0 {
			return fmt.Errorf("non-coinbase tx has no inputs (index %d)", i)
		}

	}
	return nil
}
func (h Header) FindValidNonce(ctx context.Context) (uint32, Hash32, error) {
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
			hash := h.Hash()
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

func GenerateBlock(ctx context.Context, oldBlock Block, body Body, version uint32, difficulty uint32) (Block, error) {

	var newHeader Header
	t := time.Now()

	newHeader.Version = version
	newHeader.PrevHash = CalculateHash(oldBlock)
	newHeader.Timestamp = uint32(t.Unix())
	newHeader.MerkleRoot = body.MerkleRootHash()
	newHeader.Difficulty = difficulty
	nonce, _, err := newHeader.FindValidNonce(ctx)
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

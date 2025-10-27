package crypto

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
)

var userCount atomic.Uint32

type Blockchain struct {
	Blocks     []Block
	ChainMutex *sync.RWMutex
	UTXOSet    map[Outpoint]UTXO
	UTXOMutex  *sync.RWMutex
}

func NewBlockchain() *Blockchain {
	return &Blockchain{
		Blocks:     []Block{},
		ChainMutex: &sync.RWMutex{},
		UTXOSet:    make(map[Outpoint]UTXO),
		UTXOMutex:  &sync.RWMutex{},
	}
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

type Outpoint struct {
	TxID  Hash32 `json:"tx_id"`
	Index uint32 `json:"index"`
}
type TxInput struct {
	Prev Outpoint
	Sig  []byte
}
type TxOutput struct {
	To    Hash32
	Value uint32
}
type Transaction struct {
	Inputs  []TxInput  `json:"inputs"`
	Outputs []TxOutput `json:"outputs"`
	TxID    Hash32     `json:"tx_id"`
}

type UTXO struct {
	Out   Outpoint
	To    Hash32
	Value uint32
}

type User struct {
	Id        uint32
	Name      string
	CreatedAt uint32
	PublicKey Hash32
}

func NewUser(name string) *User {
	id := userCount.Add(1)
	created := uint32(time.Now().Unix())

	data := fmt.Sprintf("%d:%s:%d", id, name, time.Now().UnixNano())

	pk := HashString(data)

	return &User{
		Id:        id,
		Name:      name,
		CreatedAt: created,
		PublicKey: pk,
	}
}

type Transactions []Transaction

func HashBytes(bytes []byte) Hash32 {
	return sha256.Sum256(bytes)
}

func HashString(str string) Hash32 {
	return HashBytes([]byte(str))
}

func SerializeTx(tx *Transaction) []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.LittleEndian, uint32(len(tx.Inputs)))
	for _, in := range tx.Inputs {
		buf.Write(in.Prev.TxID[:])
		_ = binary.Write(&buf, binary.LittleEndian, in.Prev.Index)
		_ = binary.Write(&buf, binary.LittleEndian, uint32(len(in.Sig)))
		buf.Write(in.Sig)
	}

	_ = binary.Write(&buf, binary.LittleEndian, uint32(len(tx.Outputs)))
	for _, out := range tx.Outputs {
		_ = binary.Write(&buf, binary.LittleEndian, out.Value)
		buf.Write(out.To[:])
	}
	return buf.Bytes()
}
func (t Transaction) Hash() Hash32 {
	empty := Hash32{}
	if t.TxID != empty {
		return t.TxID
	}
	serialized := SerializeTx(&t)
	h1 := sha256.Sum256(serialized)
	h2 := sha256.Sum256(h1[:])
	return Hash32(h2)
}
func MerkleRootHash(t Transactions) Hash32 {
	if len(t) == 0 {
		return Hash32{}
	}

	hashes := make([]Hash32, 0, len(t))
	for _, tx := range t {
		hashes = append(hashes, tx.Hash())
	}

	for len(hashes) > 1 {
		if len(hashes)%2 == 1 {
			hashes = append(hashes, hashes[len(hashes)-1])
		}

		next := make([]Hash32, 0, len(hashes)/2)
		for i := 0; i < len(hashes); i += 2 {
			left := hashes[i][:]
			right := hashes[i+1][:]
			concat := append(left, right...)
			h1 := sha256.Sum256(concat)
			h2 := sha256.Sum256(h1[:])
			next = append(next, Hash32(h2))
		}
		hashes = next
	}

	return hashes[0]
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
	oldBlock := b.Blocks[len(b.Blocks)-1]
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

func ValidateTransaction(tx Transaction) error {
	if len(tx.Inputs) == 0 {
		return errors.New("tx has no inputs")
	}
	return nil
}
func (bc *Blockchain) ValidateBlock(b Block) error {
	bc.ChainMutex.RLock()
	height := len(bc.Blocks)
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

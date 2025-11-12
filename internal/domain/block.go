package domain

import (
	"bytes"
	"encoding/binary"
)

// Header represents a block header.
// No custom JSON marshaling needed - Go's default encoder handles exported fields perfectly.
// This is idiomatic Go, unlike the old blockchain package which used private fields + getters/setters.
type Header struct {
	Version    uint32 `json:"version"`
	Timestamp  uint32 `json:"timestamp"`
	PrevHash   Hash32 `json:"prev_hash"`
	MerkleRoot Hash32 `json:"merkle_root"`
	Difficulty uint32 `json:"difficulty"`
	Nonce      uint32 `json:"nonce"`
}

func (h Header) Serialize() []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.LittleEndian, h.Version)
	_ = binary.Write(&buf, binary.LittleEndian, h.PrevHash)
	_ = binary.Write(&buf, binary.LittleEndian, h.MerkleRoot)
	_ = binary.Write(&buf, binary.LittleEndian, h.Timestamp)
	_ = binary.Write(&buf, binary.LittleEndian, h.Difficulty)
	_ = binary.Write(&buf, binary.LittleEndian, h.Nonce)

	return buf.Bytes()
}

// NewHeader creates a new block header
func NewHeader(version, timestamp uint32, prevHash, merkleRoot Hash32, difficulty, nonce uint32) *Header {
	return &Header{
		Version:    version,
		Timestamp:  timestamp,
		PrevHash:   prevHash,
		MerkleRoot: merkleRoot,
		Difficulty: difficulty,
		Nonce:      nonce,
	}
}

// Body represents a block body containing transactions
type Body struct {
	Transactions []Transaction `json:"transactions"`
}

// NewBody creates a new block body
func NewBody(transactions []Transaction) *Body {
	return &Body{
		Transactions: transactions,
	}
}

// Block represents a complete block in the blockchain.
// No custom JSON marshaling needed - the standard library handles this perfectly.
type Block struct {
	Header Header `json:"header"`
	Body   Body   `json:"body"`
}

// NewBlock creates a new block
func NewBlock(header Header, body Body) *Block {
	return &Block{
		Header: header,
		Body:   body,
	}
}

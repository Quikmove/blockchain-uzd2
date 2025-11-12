package domain

import (
	"bytes"
	"encoding/binary"
)

// Outpoint references a specific output in a transaction
type Outpoint struct {
	TxID  Hash32 `json:"tx_id"`
	Index uint32 `json:"index"`
}

// TxInput represents an input to a transaction
type TxInput struct {
	Prev Outpoint `json:"prev"`
	Sig  []byte   `json:"sig"`
}

// TxOutput represents an output of a transaction
type TxOutput struct {
	To    PublicAddress `json:"to"`
	Value uint32        `json:"value"`
}

// UTXO represents an unspent transaction output
type UTXO struct {
	Outpoint Outpoint
	To       PublicAddress
	Value    uint32
}

// Transaction represents a blockchain transaction
type Transaction struct {
	TxID    Hash32     `json:"txid"`
	Inputs  []TxInput  `json:"vin"`
	Outputs []TxOutput `json:"vout"`
}

// NewTransaction creates a new transaction
func NewTransaction(inputs []TxInput, outputs []TxOutput) *Transaction {
	return &Transaction{
		Inputs:  inputs,
		Outputs: outputs,
	}
}

// IsCoinbase returns true if this is a coinbase transaction (no inputs)
func (t *Transaction) IsCoinbase() bool {
	return len(t.Inputs) == 0
}
func (t *Transaction) Serialize() []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.LittleEndian, uint32(len(t.Inputs)))
	for _, in := range t.Inputs {
		buf.Write(in.Prev.TxID[:])
		_ = binary.Write(&buf, binary.LittleEndian, in.Prev.Index)
		_ = binary.Write(&buf, binary.LittleEndian, uint32(len(in.Sig)))
		buf.Write(in.Sig)
	}

	_ = binary.Write(&buf, binary.LittleEndian, uint32(len(t.Outputs)))
	for _, out := range t.Outputs {
		_ = binary.Write(&buf, binary.LittleEndian, out.Value)
		buf.Write(out.To[:])
	}
	return buf.Bytes()
}

func (t *Transaction) SerializeWithoutSignatures() []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.LittleEndian, uint32(len(t.Inputs)))
	for _, in := range t.Inputs {
		buf.Write(in.Prev.TxID[:])
		_ = binary.Write(&buf, binary.LittleEndian, in.Prev.Index)
	}

	_ = binary.Write(&buf, binary.LittleEndian, uint32(len(t.Outputs)))
	for _, out := range t.Outputs {
		_ = binary.Write(&buf, binary.LittleEndian, out.Value)
		buf.Write(out.To[:])
	}
	return buf.Bytes()
}

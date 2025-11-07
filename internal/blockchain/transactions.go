package blockchain

import (
	"bytes"
	"encoding/binary"

	"github.com/Quikmove/blockchain-uzd2/internal/merkletree"
)

type Outpoint struct {
	TxID  Hash32 `json:"tx_id"`
	Index uint32 `json:"index"`
}
type TxInput struct {
	Prev Outpoint `json:"prev"`
	Sig  []byte   `json:"sig"`
}
type TxOutput struct {
	To    Hash32 `json:"to"`
	Value uint32 `json:"value"`
}
type Transaction struct {
	TxID    Hash32     `json:"txid"`
	Inputs  []TxInput  `json:"vin"`
	Outputs []TxOutput `json:"vout"`
}

func (t *Transaction) SignatureHash(value uint32, to Hash32, hasher Hasher) (Hash32, error) {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.LittleEndian, uint32(len(t.Inputs)))
	for _, in := range t.Inputs {
		buf.Write(in.Prev.TxID[:])
		_ = binary.Write(&buf, binary.LittleEndian, in.Prev.Index)
	}
	_ = binary.Write(&buf, binary.LittleEndian, uint32(len(t.Outputs)))
	for _, out := range t.Outputs {
		buf.Write(out.To[:])
		_ = binary.Write(&buf, binary.LittleEndian, out.Value)
	}
	_ = binary.Write(&buf, binary.LittleEndian, value)
	buf.Write(to[:])

	h1 := hasher.Hash(buf.Bytes())
	h2 := hasher.Hash(h1)

	var hash Hash32
	copy(hash[:], h2)
	return hash, nil
}

type UTXO struct {
	Out   Outpoint
	To    Hash32
	Value uint32
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
func (t *Transaction) Hash(hasher Hasher) (Hash32, error) {
	empty := Hash32{}
	if t.TxID != empty {
		return t.TxID, nil
	}
	serialized := SerializeTx(t)
	h1 := hasher.Hash(serialized)
	h2 := hasher.Hash(h1)
	var hash Hash32
	copy(hash[:], h2)
	return hash, nil
}
func merkleRootHash(t Transactions, hasher Hasher) Hash32 {
	if len(t) == 0 {
		return Hash32{}
	}
	hashes := make([]merkletree.Hash32, 0, len(t))
	for _, tx := range t {
		h, err := tx.Hash(hasher)
		if err != nil {
			panic(err)
		}
		var mh merkletree.Hash32
		copy(mh[:], h[:])
		hashes = append(hashes, mh)
	}
	mt := merkletree.NewMerkleTree(hashes)
	return Hash32(mt.Root.Val)
}

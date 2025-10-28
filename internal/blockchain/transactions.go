package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"

	"github.com/Quikmove/blockchain-uzd2/internal/merkletree"
)

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
	TxID    Hash32     `json:"txid"`
	Inputs  []TxInput  `json:"vin"`
	Outputs []TxOutput `json:"vout"`
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
func merkleRootHash(t Transactions) Hash32 {
	if len(t) == 0 {
		return Hash32{}
	}
	hashes := make([]merkletree.Hash32, 0, len(t))
	for _, tx := range t {
		h := tx.Hash()
		var mh merkletree.Hash32
		copy(mh[:], h[:])
		hashes = append(hashes, mh)
	}
	mt := merkletree.NewMerkleTree(hashes)
	return Hash32(mt.Root.Val)
}

// Deprecated: use merkleRootHash instead
func _merkleRootHash(t Transactions) Hash32 {
	if len(t) == 0 {
		return Hash32{}
	}
	hashes := make([]Hash32, 0, len(t))
	for _, tx := range t {
		hashes = append(hashes, tx.Hash())
	}
	return _MerkleRootHash(hashes)
}

// Deprecated: use merkleRootHash instead
func _MerkleRootHash(hashes []Hash32) Hash32 {
	if len(hashes) == 0 {
		return Hash32{}
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
func ValidateTransaction(tx Transaction) error {
	if len(tx.Inputs) == 0 {
		return errors.New("tx has no inputs")
	}

	return nil
}

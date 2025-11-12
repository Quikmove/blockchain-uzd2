package blockchain

import (
	"bytes"
	"encoding/binary"

	c "github.com/Quikmove/blockchain-uzd2/internal/crypto"
	d "github.com/Quikmove/blockchain-uzd2/internal/domain"
	"github.com/Quikmove/blockchain-uzd2/internal/merkletree"
)

func SignatureHash(t d.Transaction, value uint32, to []byte, hasher c.Hasher) d.Hash32 {
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
	h2 := hasher.Hash(h1[:])

	return h2
}

func merkleRootHash(t Transactions, hasher c.Hasher) d.Hash32 {
	if len(t) == 0 {
		return d.Hash32{}
	}
	hashes := make([]d.Hash32, 0, len(t))
	for _, tx := range t {
		h := hasher.Hash(tx.Serialize())
		var mh d.Hash32
		copy(mh[:], h[:])
		hashes = append(hashes, mh)
	}
	mt := merkletree.NewMerkleTree(hashes)
	return mt.Root.Val
}

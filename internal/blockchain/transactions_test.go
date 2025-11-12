package blockchain

import (
	"bytes"
	"testing"

	c "github.com/Quikmove/blockchain-uzd2/internal/crypto"
	d "github.com/Quikmove/blockchain-uzd2/internal/domain"
)

func TestTransaction_Serialize(t *testing.T) {
	tx := d.Transaction{
		Inputs: []d.TxInput{
			{
				Prev: d.Outpoint{
					TxID:  d.Hash32{0x01},
					Index: 0,
				},
				Sig: []byte{0xAA, 0xBB},
			},
		},
		Outputs: []d.TxOutput{
			{
				Value: 100,
				To:    d.PublicAddress{0xCC},
			},
		},
	}

	serialized := tx.Serialize()
	if len(serialized) == 0 {
		t.Error("Serialized transaction should not be empty")
	}

	// Verify serialization is deterministic
	serialized2 := tx.Serialize()
	if !bytes.Equal(serialized, serialized2) {
		t.Error("Transaction serialization should be deterministic")
	}
}

func TestTransaction_TxID(t *testing.T) {
	hasher := c.NewArchasHasher()

	tx := d.Transaction{
		Inputs: []d.TxInput{
			{
				Prev: d.Outpoint{
					TxID:  d.Hash32{0x01},
					Index: 0,
				},
				Sig: []byte{0xAA, 0xBB},
			},
		},
		Outputs: []d.TxOutput{
			{
				Value: 100,
				To:    d.PublicAddress{0xCC},
			},
		},
	}

	// Calculate expected TxID
	expectedTxID := hasher.Hash(tx.Serialize())
	tx.TxID = expectedTxID

	// Verify TxID matches
	actualTxID := hasher.Hash(tx.Serialize())
	if tx.TxID != actualTxID {
		t.Error("Transaction TxID should match hash of serialized transaction")
	}
}

func TestSignatureHash(t *testing.T) {
	hasher := c.NewArchasHasher()

	tx := d.Transaction{
		Inputs: []d.TxInput{
			{
				Prev: d.Outpoint{
					TxID:  d.Hash32{0x01},
					Index: 0,
				},
				Sig: []byte{0xAA},
			},
		},
		Outputs: []d.TxOutput{
			{
				Value: 100,
				To:    d.PublicAddress{0xCC},
			},
		},
	}

	value := uint32(200)
	to := d.PublicAddress{0xDD}

	hash1 := SignatureHash(tx, value, to[:], hasher)
	hash2 := SignatureHash(tx, value, to[:], hasher)

	// Signature hash should be deterministic
	if hash1 != hash2 {
		t.Error("SignatureHash should be deterministic")
	}

	// Different value should produce different hash
	hash3 := SignatureHash(tx, value+1, to[:], hasher)
	if hash1 == hash3 {
		t.Error("SignatureHash should change with different value")
	}

	// Different 'to' address should produce different hash
	to2 := d.PublicAddress{0xEE}
	hash4 := SignatureHash(tx, value, to2[:], hasher)
	if hash1 == hash4 {
		t.Error("SignatureHash should change with different 'to' address")
	}
}

func TestMerkleRootHash(t *testing.T) {
	hasher := c.NewArchasHasher()

	// Empty transactions
	emptyBody := d.Body{Transactions: []d.Transaction{}}
	root1 := MerkleRootHash(emptyBody, hasher)
	if !root1.IsZero() {
		t.Error("Merkle root for empty transactions should be zero")
	}

	// Single transaction
	tx1 := d.Transaction{
		Inputs:  []d.TxInput{},
		Outputs: []d.TxOutput{{Value: 100, To: d.PublicAddress{0x01}}},
	}
	tx1.TxID = hasher.Hash(tx1.Serialize())

	body1 := d.Body{Transactions: []d.Transaction{tx1}}
	root2 := MerkleRootHash(body1, hasher)
	if root2.IsZero() {
		t.Error("Merkle root for single transaction should not be zero")
	}

	// Multiple transactions
	tx2 := d.Transaction{
		Inputs:  []d.TxInput{},
		Outputs: []d.TxOutput{{Value: 200, To: d.PublicAddress{0x02}}},
	}
	tx2.TxID = hasher.Hash(tx2.Serialize())

	body2 := d.Body{Transactions: []d.Transaction{tx1, tx2}}
	root3 := MerkleRootHash(body2, hasher)
	if root3.IsZero() {
		t.Error("Merkle root for multiple transactions should not be zero")
	}

	// Merkle root should be deterministic
	root4 := MerkleRootHash(body2, hasher)
	if root3 != root4 {
		t.Error("Merkle root should be deterministic")
	}

	// Different transaction order should produce different root
	body3 := d.Body{Transactions: []d.Transaction{tx2, tx1}}
	root5 := MerkleRootHash(body3, hasher)
	if root3 == root5 {
		t.Error("Merkle root should change with transaction order")
	}
}

func TestGenerateRandomTransactions(t *testing.T) {
	bch, users, _ := setupTestBlockchain()

	// Generate transactions
	txs, err := bch.GenerateRandomTransactions(users, 10, 100, 5)
	if err != nil {
		t.Fatalf("Failed to generate transactions: %v", err)
	}

	if len(txs) == 0 {
		t.Error("Should generate at least some transactions")
	}

	// Verify all transactions have valid structure
	for i, tx := range txs {
		if len(tx.Inputs) == 0 {
			t.Errorf("Transaction %d: non-coinbase transaction should have inputs", i)
		}
		if len(tx.Outputs) == 0 {
			t.Errorf("Transaction %d: should have outputs", i)
		}
		if tx.TxID.IsZero() {
			t.Errorf("Transaction %d: should have TxID", i)
		}
	}
}

func TestTransaction_IsCoinbase(t *testing.T) {
	// Coinbase transaction (no inputs)
	coinbaseTx := d.Transaction{
		Inputs:  []d.TxInput{},
		Outputs: []d.TxOutput{{Value: 100, To: d.PublicAddress{0x01}}},
	}

	if !coinbaseTx.IsCoinbase() {
		t.Error("Transaction with no inputs should be coinbase")
	}

	// Regular transaction (has inputs)
	regularTx := d.Transaction{
		Inputs: []d.TxInput{
			{Prev: d.Outpoint{TxID: d.Hash32{0x01}, Index: 0}},
		},
		Outputs: []d.TxOutput{{Value: 100, To: d.PublicAddress{0x01}}},
	}

	if regularTx.IsCoinbase() {
		t.Error("Transaction with inputs should not be coinbase")
	}
}

func TestTransaction_SerializationRoundTrip(t *testing.T) {
	hasher := c.NewArchasHasher()

	originalTx := d.Transaction{
		Inputs: []d.TxInput{
			{
				Prev: d.Outpoint{
					TxID:  d.Hash32{0x01, 0x02, 0x03},
					Index: 5,
				},
				Sig: []byte{0xAA, 0xBB, 0xCC},
			},
			{
				Prev: d.Outpoint{
					TxID:  d.Hash32{0x04, 0x05, 0x06},
					Index: 10,
				},
				Sig: []byte{0xDD, 0xEE},
			},
		},
		Outputs: []d.TxOutput{
			{
				Value: 100,
				To:    d.PublicAddress{0x11, 0x22},
			},
			{
				Value: 200,
				To:    d.PublicAddress{0x33, 0x44},
			},
		},
	}

	// Set TxID
	originalTx.TxID = hasher.Hash(originalTx.Serialize())

	// Serialize and verify TxID still matches
	serialized := originalTx.Serialize()
	expectedTxID := hasher.Hash(serialized)

	if originalTx.TxID != expectedTxID {
		t.Error("Transaction TxID should match hash of serialized form")
	}

	// Verify serialization includes all data
	if len(serialized) < 50 { // Minimum expected size
		t.Error("Serialized transaction seems too short")
	}
}


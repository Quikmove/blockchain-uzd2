package blockchain

import (
	"context"
	"testing"
	"time"

	c "github.com/Quikmove/blockchain-uzd2/internal/crypto"
	d "github.com/Quikmove/blockchain-uzd2/internal/domain"
)

func TestAddBlock_InvalidMerkleRoot(t *testing.T) {
	bch, users, cfg := setupTestBlockchain()
	hasher := c.NewArchasHasher()

	tx := d.Transaction{
		Inputs:  []d.TxInput{},
		Outputs: []d.TxOutput{{Value: 100, To: users[0].PublicAddress}},
	}
	tx.TxID = hasher.Hash(tx.Serialize())

	body := d.Body{Transactions: []d.Transaction{tx}}

	// Create block with wrong merkle root
	block := d.Block{
		Header: d.Header{
			Version:    cfg.Version,
			Timestamp:  uint32(time.Now().Unix()),
			MerkleRoot: d.Hash32{}, // Wrong merkle root
			Difficulty: cfg.Difficulty,
			Nonce:      1,
		},
		Body: body,
	}

	err := bch.AddBlock(block)
	if err == nil {
		t.Error("Expected error for invalid merkle root")
	}
}

func TestAddBlock_InvalidPrevHash(t *testing.T) {
	bch, users, cfg := setupTestBlockchain()
	hasher := c.NewArchasHasher()
	txSigner := c.NewTransactionSigner()

	// Create a valid transaction
	utxos := bch.GetUTXOsForAddress(users[0].PublicAddress)
	if len(utxos) == 0 {
		t.Fatal("No UTXOs available")
	}

	utxo := utxos[0]
	tx := d.Transaction{
		Inputs:  []d.TxInput{{Prev: utxo.Outpoint}},
		Outputs: []d.TxOutput{{Value: 50, To: users[1].PublicAddress}},
	}

	hashToSign := SignatureHash(tx, utxo.Value, utxo.To[:], hasher)
	sig := txSigner.SignTransaction(hashToSign[:], users[0].GetPrivateKeyObject())
	tx.Inputs[0].Sig = sig[:]
	tx.TxID = hasher.Hash(tx.Serialize())

	body := d.Body{Transactions: []d.Transaction{tx}}
	merkleRoot := MerkleRootHash(body, hasher)

	// Create block with wrong previous hash
	block := d.Block{
		Header: d.Header{
			Version:    cfg.Version,
			Timestamp:  uint32(time.Now().Unix()),
			PrevHash:   d.Hash32{0xFF}, // Wrong previous hash
			MerkleRoot: merkleRoot,
			Difficulty: cfg.Difficulty,
			Nonce:      1,
		},
		Body: body,
	}

	err := bch.AddBlock(block)
	if err == nil {
		t.Error("Expected error for invalid previous hash")
	}
}

func TestAddBlock_InvalidDifficulty(t *testing.T) {
	bch, users, cfg := setupTestBlockchain()
	hasher := c.NewArchasHasher()
	txSigner := c.NewTransactionSigner()

	utxos := bch.GetUTXOsForAddress(users[0].PublicAddress)
	if len(utxos) == 0 {
		t.Fatal("No UTXOs available")
	}

	utxo := utxos[0]
	tx := d.Transaction{
		Inputs:  []d.TxInput{{Prev: utxo.Outpoint}},
		Outputs: []d.TxOutput{{Value: 50, To: users[1].PublicAddress}},
	}

	hashToSign := SignatureHash(tx, utxo.Value, utxo.To[:], hasher)
	sig := txSigner.SignTransaction(hashToSign[:], users[0].GetPrivateKeyObject())
	tx.Inputs[0].Sig = sig[:]
	tx.TxID = hasher.Hash(tx.Serialize())

	body := d.Body{Transactions: []d.Transaction{tx}}
	merkleRoot := MerkleRootHash(body, hasher)

	latestBlock, _ := bch.GetLatestBlock()
	prevHash := bch.CalculateHash(latestBlock)

	// Create block with hash that doesn't meet difficulty
	block := d.Block{
		Header: d.Header{
			Version:    cfg.Version,
			Timestamp:  uint32(time.Now().Unix()),
			PrevHash:   prevHash,
			MerkleRoot: merkleRoot,
			Difficulty: 10, // High difficulty
			Nonce:      1, // Nonce that doesn't meet difficulty
		},
		Body: body,
	}

	err := bch.AddBlock(block)
	if err == nil {
		t.Error("Expected error for hash not meeting difficulty")
	}
}

func TestAddBlock_ValidBlock(t *testing.T) {
	bch, users, cfg := setupTestBlockchain()
	hasher := c.NewArchasHasher()
	txSigner := c.NewTransactionSigner()

	utxos := bch.GetUTXOsForAddress(users[0].PublicAddress)
	if len(utxos) == 0 {
		t.Fatal("No UTXOs available")
	}

	utxo := utxos[0]
	// Use half the UTXO value to ensure we don't exceed it
	outputValue := utxo.Value / 2
	if outputValue == 0 {
		outputValue = utxo.Value // Use full value if half would be zero
	}
	if outputValue > utxo.Value {
		t.Fatalf("Test setup error: outputValue %d exceeds UTXO value %d", outputValue, utxo.Value)
	}

	tx := d.Transaction{
		Inputs:  []d.TxInput{{Prev: utxo.Outpoint}},
		Outputs: []d.TxOutput{{Value: outputValue, To: users[1].PublicAddress}},
	}

	hashToSign := SignatureHash(tx, utxo.Value, utxo.To[:], hasher)
	sig := txSigner.SignTransaction(hashToSign[:], users[0].GetPrivateKeyObject())
	tx.Inputs[0].Sig = sig[:]
	tx.TxID = hasher.Hash(tx.Serialize())

	body := d.Body{Transactions: []d.Transaction{tx}}
	merkleRoot := MerkleRootHash(body, hasher)

	latestBlock, _ := bch.GetLatestBlock()
	prevHash := bch.CalculateHash(latestBlock)

	// Generate valid block
	block, err := bch.GenerateBlock(context.Background(), body, cfg.Version, cfg.Difficulty)
	if err != nil {
		t.Fatalf("Failed to generate block: %v", err)
	}

	// Verify block structure
	if block.Header.MerkleRoot != merkleRoot {
		t.Error("Merkle root mismatch")
	}
	if block.Header.PrevHash != prevHash {
		t.Error("Previous hash mismatch")
	}

	// Add block (should succeed)
	err = bch.AddBlock(block)
	if err != nil {
		t.Errorf("Unexpected error adding valid block: %v", err)
	}

	// Verify block was added
	if bch.Len() != 2 {
		t.Errorf("Expected blockchain length 2, got %d", bch.Len())
	}
}

func TestBlockchain_ChainIntegrity(t *testing.T) {
	bch, users, cfg := setupTestBlockchain()
	hasher := c.NewArchasHasher()
	txSigner := c.NewTransactionSigner()

	// Add multiple blocks
	for i := 0; i < 3; i++ {
		utxos := bch.GetUTXOsForAddress(users[i%len(users)].PublicAddress)
		if len(utxos) == 0 {
			continue
		}

		utxo := utxos[0]
		// Use half the UTXO value to ensure we don't exceed it
		outputValue := utxo.Value / 2
		if outputValue == 0 {
			outputValue = utxo.Value // Use full value if half would be zero
		}
		if outputValue > utxo.Value {
			t.Fatalf("Test setup error: outputValue %d exceeds UTXO value %d", outputValue, utxo.Value)
		}

		tx := d.Transaction{
			Inputs:  []d.TxInput{{Prev: utxo.Outpoint}},
			Outputs: []d.TxOutput{{Value: outputValue, To: users[(i+1)%len(users)].PublicAddress}},
		}

		hashToSign := SignatureHash(tx, utxo.Value, utxo.To[:], hasher)
		sig := txSigner.SignTransaction(hashToSign[:], users[i%len(users)].GetPrivateKeyObject())
		tx.Inputs[0].Sig = sig[:]
		tx.TxID = hasher.Hash(tx.Serialize())

		body := d.Body{Transactions: []d.Transaction{tx}}
		block, err := bch.GenerateBlock(context.Background(), body, cfg.Version, cfg.Difficulty)
		if err != nil {
			t.Fatalf("Failed to generate block %d: %v", i, err)
		}

		err = bch.AddBlock(block)
		if err != nil {
			t.Fatalf("Failed to add block %d: %v", i, err)
		}
	}

	// Verify chain integrity
	blocks := bch.Blocks()
	for i := 1; i < len(blocks); i++ {
		prevBlock := blocks[i-1]
		currentBlock := blocks[i]

		prevHash := bch.CalculateHash(prevBlock)
		if prevHash != currentBlock.Header.PrevHash {
			t.Errorf("Block %d: previous hash mismatch", i)
		}

		hash := bch.CalculateHash(currentBlock)
		if !IsHashValid(hash, currentBlock.Header.Difficulty) {
			t.Errorf("Block %d: hash does not meet difficulty", i)
		}
	}
}

func TestIsHashValid(t *testing.T) {
	hasher := c.NewArchasHasher()

	tests := []struct {
		name       string
		difficulty uint32
		nonce      uint32
		expectValid bool
	}{
		{"Zero difficulty always valid", 0, 0, true},
		{"Low difficulty", 1, 0, false}, // Will need to find valid nonce
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := d.Header{
				Version:    1,
				Timestamp:  uint32(time.Now().Unix()),
				Difficulty: tt.difficulty,
				Nonce:      tt.nonce,
			}

			hash := hasher.Hash(header.Serialize())
			valid := IsHashValid(hash, tt.difficulty)

			if valid != tt.expectValid {
				t.Errorf("IsHashValid() = %v, want %v", valid, tt.expectValid)
			}
		})
	}
}

func TestCalculateHash(t *testing.T) {
	bch, _, _ := setupTestBlockchain()

	block, err := bch.GetLatestBlock()
	if err != nil {
		t.Fatalf("Failed to get latest block: %v", err)
	}

	hash := bch.CalculateHash(block)
	if hash.IsZero() {
		t.Error("Block hash should not be zero")
	}

	// Hash should be deterministic
	hash2 := bch.CalculateHash(block)
	if hash != hash2 {
		t.Error("Block hash should be deterministic")
	}
}


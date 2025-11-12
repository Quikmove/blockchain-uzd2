package blockchain

import (
	"testing"
	"time"

	"github.com/Quikmove/blockchain-uzd2/internal/config"
	c "github.com/Quikmove/blockchain-uzd2/internal/crypto"
	d "github.com/Quikmove/blockchain-uzd2/internal/domain"
)

func setupTestBlockchain() (*Blockchain, []d.User, *config.Config) {
	hasher := c.NewArchasHasher()
	txSigner := c.NewTransactionSigner()
	cfg := &config.Config{
		Version:    1,
		Difficulty: 1, // Low difficulty for faster tests
	}
	keyGen := c.NewKeyGenerator()
	userGen := NewUserGeneratorService(keyGen)
	users := userGen.GenerateUsers([]string{"Alice", "Bob", "Charlie"}, 3)
	// Use fixed, large amounts to ensure UTXOs are sufficient for testing
	bch := InitBlockchainWithFunds(100000, 100000, users, cfg, hasher, txSigner)
	return bch, users, cfg
}

func TestValidateBlock_EmptyTransactions(t *testing.T) {
	bch, _, _ := setupTestBlockchain()
	block := d.Block{
		Header: d.Header{
			Version:    1,
			Timestamp:  uint32(time.Now().Unix()),
			Difficulty: 1,
		},
		Body: d.Body{
			Transactions: []d.Transaction{},
		},
	}

	err := bch.ValidateBlock(block)
	if err == nil {
		t.Error("Expected error for block with no transactions")
	}
}

func TestValidateBlock_MerkleRootMismatch(t *testing.T) {
	bch, users, cfg := setupTestBlockchain()
	hasher := c.NewArchasHasher()

	// Create a transaction
	tx := d.Transaction{
		Inputs:  []d.TxInput{},
		Outputs: []d.TxOutput{{Value: 100, To: users[0].PublicAddress}},
	}
	tx.TxID = hasher.Hash(tx.Serialize())

	body := d.Body{Transactions: []d.Transaction{tx}}
	_ = MerkleRootHash(body, hasher) // Compute but don't use

	// Create block with wrong merkle root
	block := d.Block{
		Header: d.Header{
			Version:     cfg.Version,
			Timestamp:   uint32(time.Now().Unix()),
			MerkleRoot:  d.Hash32{}, // Wrong merkle root
			Difficulty:  cfg.Difficulty,
			Nonce:       1,
		},
		Body: body,
	}

	err := bch.ValidateBlock(block)
	if err == nil {
		t.Error("Expected error for merkle root mismatch")
	}
	if err != nil && err.Error()[:15] != "merkle root mis" {
		t.Errorf("Expected merkle root mismatch error, got: %v", err)
	}
}

func TestValidateBlock_TransactionIDMismatch(t *testing.T) {
	bch, users, cfg := setupTestBlockchain()
	hasher := c.NewArchasHasher()

	// Create a transaction with wrong TxID
	tx := d.Transaction{
		TxID:    d.Hash32{}, // Wrong TxID
		Inputs:  []d.TxInput{},
		Outputs: []d.TxOutput{{Value: 100, To: users[0].PublicAddress}},
	}

	body := d.Body{Transactions: []d.Transaction{tx}}
	merkleRoot := MerkleRootHash(body, hasher)

	block := d.Block{
		Header: d.Header{
			Version:    cfg.Version,
			Timestamp:  uint32(time.Now().Unix()),
			MerkleRoot: merkleRoot,
			Difficulty: cfg.Difficulty,
			Nonce:      1,
		},
		Body: body,
	}

	err := bch.ValidateBlock(block)
	if err == nil {
		t.Error("Expected error for transaction ID mismatch")
	}
}

func TestValidateBlock_TimestampTooFarInFuture(t *testing.T) {
	bch, users, cfg := setupTestBlockchain()
	hasher := c.NewArchasHasher()

	tx := d.Transaction{
		Inputs:  []d.TxInput{},
		Outputs: []d.TxOutput{{Value: 100, To: users[0].PublicAddress}},
	}
	tx.TxID = hasher.Hash(tx.Serialize())

	body := d.Body{Transactions: []d.Transaction{tx}}
	merkleRoot := MerkleRootHash(body, hasher)

	// Timestamp 3 hours in future (more than 2 hour limit)
	futureTime := uint32(time.Now().Unix()) + 10800
	block := d.Block{
		Header: d.Header{
			Version:    cfg.Version,
			Timestamp:  futureTime,
			MerkleRoot: merkleRoot,
			Difficulty: cfg.Difficulty,
			Nonce:      1,
		},
		Body: body,
	}

	err := bch.ValidateBlock(block)
	if err == nil {
		t.Error("Expected error for timestamp too far in future")
	}
}

func TestValidateBlock_ValidGenesisBlock(t *testing.T) {
	// Create a new empty blockchain for genesis block test
	hasher := c.NewArchasHasher()
	txSigner := c.NewTransactionSigner()
	bch := NewBlockchain(hasher, txSigner)
	
	users, cfg := func() ([]d.User, *config.Config) {
		keyGen := c.NewKeyGenerator()
		userGen := NewUserGeneratorService(keyGen)
		users := userGen.GenerateUsers([]string{"Alice"}, 1)
		cfg := &config.Config{
			Version:    1,
			Difficulty: 0, // Zero difficulty for genesis block test
		}
		return users, cfg
	}()

	tx := d.Transaction{
		Inputs:  []d.TxInput{},
		Outputs: []d.TxOutput{{Value: 100, To: users[0].PublicAddress}},
	}
	tx.TxID = hasher.Hash(tx.Serialize())

	body := d.Body{Transactions: []d.Transaction{tx}}
	merkleRoot := MerkleRootHash(body, hasher)

	block := d.Block{
		Header: d.Header{
			Version:    cfg.Version,
			Timestamp:  uint32(time.Now().Unix()),
			MerkleRoot: merkleRoot,
			Difficulty: cfg.Difficulty,
			Nonce:      1,
		},
		Body: body,
	}

	err := bch.ValidateBlock(block)
	if err != nil {
		t.Errorf("Unexpected error for valid genesis block: %v", err)
	}
}

func TestValidateBlockTransactions_DoubleSpend(t *testing.T) {
	bch, users, cfg := setupTestBlockchain()
	hasher := c.NewArchasHasher()
	txSigner := c.NewTransactionSigner()

	// Get UTXOs for first user
	utxos := bch.GetUTXOsForAddress(users[0].PublicAddress)
	if len(utxos) == 0 {
		t.Fatal("No UTXOs available for testing")
	}

	utxo := utxos[0]

	// Create transaction spending the same UTXO twice
	tx := d.Transaction{
		Inputs: []d.TxInput{
			{Prev: utxo.Outpoint},
			{Prev: utxo.Outpoint}, // Same input twice
		},
		Outputs: []d.TxOutput{{Value: 50, To: users[1].PublicAddress}},
	}

	// Sign both inputs (even though it's a double spend)
	hashToSign := SignatureHash(tx, utxo.Value, utxo.To[:], hasher)
	sig := txSigner.SignTransaction(hashToSign[:], users[0].GetPrivateKeyObject())
	tx.Inputs[0].Sig = sig[:]
	tx.Inputs[1].Sig = sig[:]

	tx.TxID = hasher.Hash(tx.Serialize())

	body := d.Body{Transactions: []d.Transaction{tx}}
	merkleRoot := MerkleRootHash(body, hasher)

	block := d.Block{
		Header: d.Header{
			Version:    cfg.Version,
			Timestamp:  uint32(time.Now().Unix()),
			MerkleRoot: merkleRoot,
			Difficulty: cfg.Difficulty,
			Nonce:      1,
		},
		Body: body,
	}

	err := bch.ValidateBlockTransactions(block, users)
	if err == nil {
		t.Error("Expected error for double spend")
	}
	if err != nil && err != d.ErrDoubleSpend {
		t.Errorf("Expected ErrDoubleSpend, got: %v", err)
	}
}

func TestValidateBlockTransactions_UTXONotFound(t *testing.T) {
	bch, users, cfg := setupTestBlockchain()
	hasher := c.NewArchasHasher()

	// Create transaction with non-existent UTXO
	fakeOutpoint := d.Outpoint{
		TxID:  d.Hash32{0xFF},
		Index: 999,
	}

	tx := d.Transaction{
		Inputs: []d.TxInput{
			{Prev: fakeOutpoint, Sig: []byte{0x01, 0x02}},
		},
		Outputs: []d.TxOutput{{Value: 50, To: users[1].PublicAddress}},
	}
	tx.TxID = hasher.Hash(tx.Serialize())

	body := d.Body{Transactions: []d.Transaction{tx}}
	merkleRoot := MerkleRootHash(body, hasher)

	block := d.Block{
		Header: d.Header{
			Version:    cfg.Version,
			Timestamp:  uint32(time.Now().Unix()),
			MerkleRoot: merkleRoot,
			Difficulty: cfg.Difficulty,
			Nonce:      1,
		},
		Body: body,
	}

	err := bch.ValidateBlockTransactions(block, users)
	if err == nil {
		t.Error("Expected error for UTXO not found")
	}
	if err != nil && err != d.ErrUTXONotFound {
		t.Errorf("Expected ErrUTXONotFound, got: %v", err)
	}
}

func TestValidateBlockTransactions_InvalidSignature(t *testing.T) {
	bch, users, cfg := setupTestBlockchain()
	hasher := c.NewArchasHasher()

	// Get UTXOs for first user
	utxos := bch.GetUTXOsForAddress(users[0].PublicAddress)
	if len(utxos) == 0 {
		t.Fatal("No UTXOs available for testing")
	}

	utxo := utxos[0]

	// Create transaction with invalid signature (wrong private key)
	input := d.TxInput{
		Prev: utxo.Outpoint,
		Sig:  []byte{0x01, 0x02, 0x03}, // Invalid signature
	}

	tx := d.Transaction{
		Inputs:  []d.TxInput{input},
		Outputs: []d.TxOutput{{Value: 50, To: users[1].PublicAddress}},
	}
	tx.TxID = hasher.Hash(tx.Serialize())

	body := d.Body{Transactions: []d.Transaction{tx}}
	merkleRoot := MerkleRootHash(body, hasher)

	block := d.Block{
		Header: d.Header{
			Version:    cfg.Version,
			Timestamp:  uint32(time.Now().Unix()),
			MerkleRoot: merkleRoot,
			Difficulty: cfg.Difficulty,
			Nonce:      1,
		},
		Body: body,
	}

	err := bch.ValidateBlockTransactions(block, users)
	if err == nil {
		t.Error("Expected error for invalid signature")
	}
}

func TestValidateBlockTransactions_ValidTransaction(t *testing.T) {
	bch, users, cfg := setupTestBlockchain()
	hasher := c.NewArchasHasher()
	txSigner := c.NewTransactionSigner()

	// Get UTXOs for first user
	utxos := bch.GetUTXOsForAddress(users[0].PublicAddress)
	if len(utxos) == 0 {
		t.Fatal("No UTXOs available for testing")
	}

	utxo := utxos[0]

	// Create valid transaction - use a value less than UTXO value to leave room for change
	// Or use the full UTXO value if it's small
	outputValue := utxo.Value / 2
	if outputValue == 0 {
		outputValue = utxo.Value // Use full value if half would be zero
	}
	if outputValue > utxo.Value {
		t.Fatalf("Test setup error: outputValue %d exceeds UTXO value %d", outputValue, utxo.Value)
	}

	// Create valid transaction
	tx := d.Transaction{
		Inputs:  []d.TxInput{{Prev: utxo.Outpoint}},
		Outputs: []d.TxOutput{{Value: outputValue, To: users[1].PublicAddress}},
	}

	// Sign the transaction
	hashToSign := SignatureHash(tx, utxo.Value, utxo.To[:], hasher)
	sig := txSigner.SignTransaction(hashToSign[:], users[0].GetPrivateKeyObject())
	tx.Inputs[0].Sig = sig[:]

	// Set transaction ID
	tx.TxID = hasher.Hash(tx.Serialize())

	body := d.Body{Transactions: []d.Transaction{tx}}
	merkleRoot := MerkleRootHash(body, hasher)

	block := d.Block{
		Header: d.Header{
			Version:    cfg.Version,
			Timestamp:  uint32(time.Now().Unix()),
			MerkleRoot: merkleRoot,
			Difficulty: cfg.Difficulty,
			Nonce:      1,
		},
		Body: body,
	}

	err := bch.ValidateBlockTransactions(block, users)
	if err != nil {
		t.Errorf("Unexpected error for valid transaction: %v", err)
	}
}

func TestValidateBlockTransactions_OutputsExceedInputs(t *testing.T) {
	bch, users, cfg := setupTestBlockchain()
	hasher := c.NewArchasHasher()
	txSigner := c.NewTransactionSigner()

	// Get UTXOs for first user
	utxos := bch.GetUTXOsForAddress(users[0].PublicAddress)
	if len(utxos) == 0 {
		t.Fatal("No UTXOs available for testing")
	}

	utxo := utxos[0]

	// Create transaction where outputs exceed inputs
	tx := d.Transaction{
		Inputs: []d.TxInput{{Prev: utxo.Outpoint}},
		Outputs: []d.TxOutput{
			{Value: utxo.Value + 1000, To: users[1].PublicAddress}, // More than input
		},
	}

	hashToSign := SignatureHash(tx, utxo.Value, utxo.To[:], hasher)
	sig := txSigner.SignTransaction(hashToSign[:], users[0].GetPrivateKeyObject())
	tx.Inputs[0].Sig = sig[:]

	tx.TxID = hasher.Hash(tx.Serialize())

	body := d.Body{Transactions: []d.Transaction{tx}}
	merkleRoot := MerkleRootHash(body, hasher)

	block := d.Block{
		Header: d.Header{
			Version:    cfg.Version,
			Timestamp:  uint32(time.Now().Unix()),
			MerkleRoot: merkleRoot,
			Difficulty: cfg.Difficulty,
			Nonce:      1,
		},
		Body: body,
	}

	err := bch.ValidateBlockTransactions(block, users)
	if err == nil {
		t.Error("Expected error for outputs exceeding inputs")
	}
}

func TestValidateBlockTransactions_ZeroValueOutput(t *testing.T) {
	bch, users, cfg := setupTestBlockchain()
	hasher := c.NewArchasHasher()
	txSigner := c.NewTransactionSigner()

	utxos := bch.GetUTXOsForAddress(users[0].PublicAddress)
	if len(utxos) == 0 {
		t.Fatal("No UTXOs available for testing")
	}

	utxo := utxos[0]

	tx := d.Transaction{
		Inputs: []d.TxInput{{Prev: utxo.Outpoint}},
		Outputs: []d.TxOutput{
			{Value: 0, To: users[1].PublicAddress}, // Zero value
		},
	}

	hashToSign := SignatureHash(tx, utxo.Value, utxo.To[:], hasher)
	sig := txSigner.SignTransaction(hashToSign[:], users[0].GetPrivateKeyObject())
	tx.Inputs[0].Sig = sig[:]

	tx.TxID = hasher.Hash(tx.Serialize())

	body := d.Body{Transactions: []d.Transaction{tx}}
	merkleRoot := MerkleRootHash(body, hasher)

	block := d.Block{
		Header: d.Header{
			Version:    cfg.Version,
			Timestamp:  uint32(time.Now().Unix()),
			MerkleRoot: merkleRoot,
			Difficulty: cfg.Difficulty,
			Nonce:      1,
		},
		Body: body,
	}

	err := bch.ValidateBlockTransactions(block, users)
	if err == nil {
		t.Error("Expected error for zero-value output")
	}
}


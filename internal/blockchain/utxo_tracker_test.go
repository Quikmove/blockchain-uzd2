package blockchain

import (
	"testing"
)

func newTestUser(name string) User {
	return *NewUser(name)
}

func TestUTXOTracker(t *testing.T) {
	hasher := NewSHA256Hasher()
	tracker := NewUTXOTracker()

	user1 := newTestUser("Alice")
	user2 := newTestUser("Bob")

	// 1. Genesis-like block (coinbase transaction)
	tx1 := Transaction{
		Outputs: []TxOutput{{Value: 100, To: user1.PublicKey}},
	}
	tx1Hash, err := tx1.Hash(hasher)
	if err != nil {
		t.Fatalf("hash tx1: %v", err)
	}
	tx1.TxID = tx1Hash

	block1 := Block{Body: Body{Transactions: []Transaction{tx1}}}

	tracker.ScanBlock(block1, hasher)

	// Check balance of user1
	balance1 := tracker.GetBalance(user1.PublicKey)
	if balance1 != 100 {
		t.Fatalf("user1 balance = %d, want 100", balance1)
	}

	// Check UTXOs for user1
	utxos1 := tracker.GetUTXOsForAddress(user1.PublicKey)
	if len(utxos1) != 1 {
		t.Fatalf("user1 utxos len = %d, want 1", len(utxos1))
	}
	if utxos1[0].Out.TxID != tx1Hash {
		t.Fatalf("user1 utxo txid = %v, want %v", utxos1[0].Out.TxID, tx1Hash)
	}
	if utxos1[0].Out.Index != 0 {
		t.Fatalf("user1 utxo index = %d, want 0", utxos1[0].Out.Index)
	}
	if utxos1[0].Value != 100 {
		t.Fatalf("user1 utxo value = %d, want 100", utxos1[0].Value)
	}

	// 2. Spend previous UTXO, send 60 to user2, 40 change to user1
	inputUTXO := utxos1[0]
	tx2 := Transaction{
		Inputs:  []TxInput{{Prev: inputUTXO.Out}},
		Outputs: []TxOutput{{Value: 60, To: user2.PublicKey}, {Value: 40, To: user1.PublicKey}},
	}
	tx2Hash, err := tx2.Hash(hasher)
	if err != nil {
		t.Fatalf("hash tx2: %v", err)
	}
	tx2.TxID = tx2Hash

	block2 := Block{Body: Body{Transactions: []Transaction{tx2}}}

	tracker.ScanBlock(block2, hasher)

	// Old UTXO should be removed
	if _, exists := tracker.GetUTXO(inputUTXO.Out); exists {
		t.Fatalf("spent UTXO still exists")
	}

	// New balances
	balance1 = tracker.GetBalance(user1.PublicKey)
	if balance1 != 40 {
		t.Fatalf("user1 balance = %d, want 40", balance1)
	}
	balance2 := tracker.GetBalance(user2.PublicKey)
	if balance2 != 60 {
		t.Fatalf("user2 balance = %d, want 60", balance2)
	}

	// New UTXOs for user1 (change)
	utxos1 = tracker.GetUTXOsForAddress(user1.PublicKey)
	if len(utxos1) != 1 {
		t.Fatalf("user1 utxos len = %d, want 1", len(utxos1))
	}
	if utxos1[0].Out.TxID != tx2Hash || utxos1[0].Out.Index != 1 || utxos1[0].Value != 40 {
		t.Fatalf("user1 change UTXO mismatch: got {tx:%v idx:%d val:%d}", utxos1[0].Out.TxID, utxos1[0].Out.Index, utxos1[0].Value)
	}

	// New UTXOs for user2
	utxos2 := tracker.GetUTXOsForAddress(user2.PublicKey)
	if len(utxos2) != 1 {
		t.Fatalf("user2 utxos len = %d, want 1", len(utxos2))
	}
	if utxos2[0].Out.TxID != tx2Hash || utxos2[0].Out.Index != 0 || utxos2[0].Value != 60 {
		t.Fatalf("user2 UTXO mismatch: got {tx:%v idx:%d val:%d}", utxos2[0].Out.TxID, utxos2[0].Out.Index, utxos2[0].Value)
	}

	// 3. Test ScanBlockchain on a new tracker
	bc := NewBlockchain(hasher)
	// Append blocks directly to bypass other validations for this focused test
	bc.blocks = append(bc.blocks, block1, block2)

	tracker.ScanBlockchain(bc)

	// Re-verify balances after full scan
	balance1 = tracker.GetBalance(user1.PublicKey)
	if balance1 != 40 {
		t.Fatalf("user1 balance after rescan = %d, want 40", balance1)
	}
	balance2 = tracker.GetBalance(user2.PublicKey)
	if balance2 != 60 {
		t.Fatalf("user2 balance after rescan = %d, want 60", balance2)
	}
	if len(tracker.utxoSet) != 2 {
		t.Fatalf("utxoSet size after rescan = %d, want 2", len(tracker.utxoSet))
	}

	// 4. Test reset
	tracker.reset()
	if len(tracker.utxoSet) != 0 {
		t.Fatalf("utxoSet not empty after reset (size %d)", len(tracker.utxoSet))
	}
}

func TestGetUTXO(t *testing.T) {
	tracker := NewUTXOTracker()
	hasher := NewSHA256Hasher()
	user := newTestUser("Charlie")

	tx := Transaction{Outputs: []TxOutput{{Value: 50, To: user.PublicKey}}}
	txHash, err := tx.Hash(hasher)
	if err != nil {
		t.Fatalf("hash tx: %v", err)
	}
	tx.TxID = txHash

	block := Block{Body: Body{Transactions: []Transaction{tx}}}
	tracker.ScanBlock(block, hasher)

	outpoint := Outpoint{TxID: txHash, Index: 0}
	utxo, exists := tracker.GetUTXO(outpoint)
	if !exists {
		t.Fatalf("UTXO not found")
	}
	if utxo.Out != outpoint {
		t.Fatalf("outpoint mismatch: got %+v want %+v", utxo.Out, outpoint)
	}
	if utxo.Value != 50 {
		t.Fatalf("value = %d, want 50", utxo.Value)
	}
	if utxo.To != user.PublicKey {
		t.Fatalf("owner mismatch")
	}

	nonExistentOutpoint := Outpoint{TxID: Hash32{1}, Index: 0}
	if _, exists = tracker.GetUTXO(nonExistentOutpoint); exists {
		t.Fatalf("unexpectedly found non-existent UTXO")
	}
}

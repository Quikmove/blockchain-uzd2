package merkletree_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/Quikmove/blockchain-uzd2/internal/merkletree"
)

func TestEvenLeavesResult(t *testing.T) {
	names := []string{"Alice", "Bob", "Charlie", "David"}
	hashes := make([]merkletree.Hash32, len(names))
	for i, name := range names {
		hash := sha256.Sum256([]byte(name))
		doubleHash := sha256.Sum256(hash[:])
		hashes[i] = merkletree.Hash32(doubleHash)
	}
	tree := merkletree.NewMerkleTree(hashes)
	got := tree.Root.Val
	wantString := "5549d62ecdef8912a5d79d8385d1cce2f2cb6dd6fb67d1165ae3e334019632c5"
	var want merkletree.Hash32
	bytes, err := hex.DecodeString(wantString)
	if err != nil {
		t.Fatalf("Failed to decode hex string: %v", err)
	}
	copy(want[:], bytes)
	if got != want {
		t.Errorf("Merkle tree root hash = %x, want %x", got, want)
	}
}

func TestOddLeavesResult(t *testing.T) {
	names := []string{"Alice", "Bob", "Charlie", "David", "Eve"}
	hashes := make([]merkletree.Hash32, len(names))
	for i, name := range names {
		hash := sha256.Sum256([]byte(name))
		doubleHash := sha256.Sum256(hash[:])
		hashes[i] = merkletree.Hash32(doubleHash)
	}
	tree := merkletree.NewMerkleTree(hashes)
	got := tree.Root.Val
	wantString := "399f41145ce72adec5531d293e3cf0db033c9cd7ca7cf73a16c2a1f660b15976"
	var want merkletree.Hash32
	bytes, err := hex.DecodeString(wantString)
	if err != nil {
		t.Fatalf("Failed to decode hex string: %v", err)
	}
	copy(want[:], bytes)
	if got != want {
		t.Errorf("Merkle tree root hash = %x, want %x", got, want)
	}
}

package blockchain

import (
	"encoding/hex"
	"math/bits"
	"strings"
	"testing"
)

func bitDiffHex(hex1, hex2 string) (int, error) {
	bytes1, err := hex.DecodeString(hex1)
	if err != nil {
		return 0, err
	}
	bytes2, err := hex.DecodeString(hex2)
	if err != nil {
		return 0, err
	}

	diff := 0
	for i := 0; i < len(bytes1); i++ {
		diff += bits.OnesCount8(bytes1[i] ^ bytes2[i])
	}
	return diff, nil
}

func TestAvalancheSingleCharacter(t *testing.T) {
	hasher := NewArchasHasher()
	base := []byte(strings.Repeat("a", 10000))

	hash1Bytes, err := hasher.Hash(base)
	if err != nil {
		t.Fatalf("Failed to hash base string: %v", err)
	}
	hash1 := hex.EncodeToString(hash1Bytes)

	base[4321] = 'b'

	hash2Bytes, err := hasher.Hash(base)
	if err != nil {
		t.Fatalf("Failed to hash mutated string: %v", err)
	}
	hash2 := hex.EncodeToString(hash2Bytes)

	diff, err := bitDiffHex(hash1, hash2)
	if err != nil {
		t.Fatalf("Failed to calculate bit difference: %v", err)
	}

	if diff <= 120 {
		t.Errorf("Expected strong avalanche for long input; got diff %d, want > 120", diff)
	}
}

func TestAvalancheBitFlipsAcrossMessage(t *testing.T) {
	hasher := NewArchasHasher()
	base := "Hash functions should react strongly to minimal perturbations."

	originalBytes, err := hasher.Hash([]byte(base))
	if err != nil {
		t.Fatalf("Failed to hash base string: %v", err)
	}
	originalHash := hex.EncodeToString(originalBytes)

	totalDiff := 0
	samples := 0
	for i := 0; i < len(base); i++ {
		mutated := []byte(base)
		mutated[i] ^= 0x01 // Flip one bit

		mutatedHashBytes, err := hasher.Hash(mutated)
		if err != nil {
			t.Fatalf("Failed to hash mutated string at position %d: %v", i, err)
		}
		mutatedHash := hex.EncodeToString(mutatedHashBytes)

		diff, err := bitDiffHex(originalHash, mutatedHash)
		if err != nil {
			t.Fatalf("Failed to calculate bit difference at position %d: %v", i, err)
		}

		if diff <= 64 {
			t.Errorf("Weak avalanche at position %d; got diff %d, want > 64", i, diff)
		}
		totalDiff += diff
		samples++
	}
}

func TestHasherComparison(t *testing.T) {
	archasHasher := NewArchasHasher()
	sha256Hasher := NewSHA256Hasher()

	inputs := []string{
		"lietuva",
		"Lietuva",
		"Lietuva!",
		"Lietuva!!",
	}

	t.Logf("\n| %-9s | %-64s | %-64s |", "Įvestis", "AIHasher (Arkas)", "Hasher (SHA256)")
	t.Logf("|%-11s|%-66s|%-66s|", strings.Repeat("-", 11), strings.Repeat("-", 66), strings.Repeat("-", 66))

	for _, input := range inputs {
		archasHashBytes, err := archasHasher.Hash([]byte(input))
		if err != nil {
			t.Fatalf("ArchasHasher failed for input '%s': %v", input, err)
		}
		archasHash := hex.EncodeToString(archasHashBytes)

		sha256HashBytes, err := sha256Hasher.Hash([]byte(input))
		if err != nil {
			t.Fatalf("SHA256Hasher failed for input '%s': %v", input, err)
		}
		sha256Hash := hex.EncodeToString(sha256HashBytes)

		t.Logf("| %-9s | %-64s | %-64s |", input, archasHash, sha256Hash)
	}
}

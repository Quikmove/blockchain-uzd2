package domain

import (
	"encoding/json"
	"testing"
)

func TestHash32String(t *testing.T) {
	h := Hash32{
		0x1d, 0x1c, 0x1b, 0x1a, 0x19, 0x18, 0x17, 0x16,
		0x15, 0x14, 0x13, 0x12, 0x11, 0x10, 0x0f, 0x0e,
		0x0d, 0x0c, 0x0b, 0x0a, 0x09, 0x08, 0x07, 0x06,
		0x05, 0x04, 0x03, 0x02, 0x01, 0x00, 0x00, 0x00,
	}

	be := h.StringBE()
	expected_be := "0000000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d"
	if be != expected_be {
		t.Errorf("StringBE() = %s, want %s", be, expected_be)
	}

	le := h.StringLE()
	expected_le := "1d1c1b1a191817161514131211100f0e0d0c0b0a090807060504030201000000"
	if le != expected_le {
		t.Errorf("StringLE() = %s, want %s", le, expected_le)
	}

	str := h.String()
	if str != be {
		t.Errorf("String() = %s, want %s (should match StringBE)", str, be)
	}
}

func TestHash32Reverse(t *testing.T) {
	original := Hash32{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
		0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
	}

	reversed := original.Reverse()

	expected := Hash32{
		0x1f, 0x1e, 0x1d, 0x1c, 0x1b, 0x1a, 0x19, 0x18,
		0x17, 0x16, 0x15, 0x14, 0x13, 0x12, 0x11, 0x10,
		0x0f, 0x0e, 0x0d, 0x0c, 0x0b, 0x0a, 0x09, 0x08,
		0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01, 0x00,
	}

	if reversed != expected {
		t.Errorf("Reverse() = %v, want %v", reversed, expected)
	}

	// Reversing twice should give original
	doubleReversed := reversed.Reverse()
	if doubleReversed != original {
		t.Errorf("Reverse().Reverse() = %v, want %v", doubleReversed, original)
	}
}

func TestHash32JSON(t *testing.T) {
	h := &Hash32{
		0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0xf0, 0xde, 0xbc, 0x9a, 0x78, 0x56, 0x34,
		0x12, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	data, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		t.Fatalf("Unmarshal to string error = %v", err)
	}

	if str != h.StringBE() {
		t.Errorf("JSON marshaled as %s, want %s (big-endian)", str, h.StringBE())
	}

	var h2 *Hash32
	if err := json.Unmarshal(data, &h2); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if !h.Equals(h2) {
		t.Errorf("Unmarshal = %v, want %v", h2, h)
	}
}

func TestHash32JSONInStruct(t *testing.T) {
	prevHash := Hash32{
		0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0xf0, 0xde, 0xbc, 0x9a, 0x78, 0x56, 0x34,
		0x12, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	merkleRoot := Hash32{
		0x20, 0x1f, 0x1e, 0x1d, 0x1c, 0x1b, 0x1a, 0x19,
		0x18, 0x17, 0x16, 0x15, 0x14, 0x13, 0x12, 0x11,
		0x10, 0x0f, 0x0e, 0x0d, 0x0c, 0x0b, 0x0a, 0x09,
		0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01,
	}

	header := Header{
		Version:    1,
		Timestamp:  1234567890,
		PrevHash:   prevHash,
		MerkleRoot: merkleRoot,
		Difficulty: 4,
		Nonce:      42,
	}

	data, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	prevHashStr, ok := result["prev_hash"].(string)
	if !ok {
		t.Fatalf("prev_hash should be a string, got %T: %v", result["prev_hash"], result["prev_hash"])
	}

	if prevHashStr != prevHash.StringBE() {
		t.Errorf("prev_hash marshaled as %s, want %s (big-endian)", prevHashStr, prevHash.StringBE())
	}

	merkleRootStr, ok := result["merkle_root"].(string)
	if !ok {
		t.Fatalf("merkle_root should be a string, got %T: %v", result["merkle_root"], result["merkle_root"])
	}

	if merkleRootStr != merkleRoot.StringBE() {
		t.Errorf("merkle_root marshaled as %s, want %s (big-endian)", merkleRootStr, merkleRoot.StringBE())
	}

	var header2 Header
	if err := json.Unmarshal(data, &header2); err != nil {
		t.Fatalf("json.Unmarshal() to Header error = %v", err)
	}

	if !prevHash.Equals(&header2.PrevHash) {
		t.Errorf("Unmarshaled prev_hash = %v, want %v", header2.PrevHash, prevHash)
	}

	if !merkleRoot.Equals(&header2.MerkleRoot) {
		t.Errorf("Unmarshaled merkle_root = %v, want %v", header2.MerkleRoot, merkleRoot)
	}
}

func TestHash32LeadingZeros(t *testing.T) {
	h := Hash32{
		0x22, 0x11, 0x00, 0xff, 0xee, 0xdd, 0xcc, 0xbb,
		0xaa, 0x99, 0x88, 0x77, 0x66, 0x55, 0x44, 0x33,
		0x22, 0x11, 0xf0, 0xde, 0xbc, 0x9a, 0x78, 0x56,
		0x34, 0x12, 0xef, 0xcd, 0xab, 0x00, 0x00, 0x00,
	}

	le := h.StringLE()
	if le[len(le)-6:] != "000000" {
		t.Errorf("Little-endian display should show leading zeros at end, got: %s", le)
	}

	be := h.StringBE()
	if be[:6] != "000000" {
		t.Errorf("Big-endian should show leading zeros at start, got: %s", be)
	}
}

func TestHash32IsZero(t *testing.T) {
	zero := Hash32{}
	if !zero.IsZero() {
		t.Error("Zero hash should return true for IsZero()")
	}

	nonZero := Hash32{0x00, 0x00, 0x01}
	if nonZero.IsZero() {
		t.Error("Non-zero hash should return false for IsZero()")
	}
}

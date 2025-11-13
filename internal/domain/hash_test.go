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

	str := h.String()
	expected := "1d1c1b1a191817161514131211100f0e0d0c0b0a090807060504030201000000"
	if str != expected {
		t.Errorf("String() = %s, want %s", str, expected)
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

	if str != h.String() {
		t.Errorf("JSON marshaled as %s, want %s", str, h.String())
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

	if prevHashStr != prevHash.String() {
		t.Errorf("prev_hash marshaled as %s, want %s", prevHashStr, prevHash.String())
	}

	merkleRootStr, ok := result["merkle_root"].(string)
	if !ok {
		t.Fatalf("merkle_root should be a string, got %T: %v", result["merkle_root"], result["merkle_root"])
	}

	if merkleRootStr != merkleRoot.String() {
		t.Errorf("merkle_root marshaled as %s, want %s", merkleRootStr, merkleRoot.String())
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

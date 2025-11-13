package domain

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// Hash32 represents a 32-byte hash value used throughout the blockchain.
type Hash32 [32]byte

// String returns the hexadecimal string representation of the hash in little-endian format.
// This follows Bitcoin's convention where hashes are displayed reversed.
// For example, a hash computed as [0x01, 0x02, 0x03...] displays as "...030201"
func (h *Hash32) String() string {
	return h.StringLE()
}

// StringBE returns the hash as a big-endian hex string (natural byte order).
// This is the internal representation - bytes in the order they were computed.
func (h *Hash32) StringBE() string {
	return hex.EncodeToString(h[:])
}

// StringLE returns the hash as a little-endian hex string (reversed byte order).
// This is the standard display format for blockchain hashes, following Bitcoin convention.
// Most block explorers and tools display hashes this way.
func (h *Hash32) StringLE() string {
	reversed := h.Reverse()
	return hex.EncodeToString(reversed[:])
}

// Reverse returns a new Hash32 with bytes in reversed order.
// This is used for converting between internal representation and display format.
func (h *Hash32) Reverse() Hash32 {
	var reversed Hash32
	for i := 0; i < 32; i++ {
		reversed[i] = h[31-i]
	}
	return reversed
}

// MarshalJSON implements json.Marshaler.
// Hashes are marshaled as little-endian hex strings (reversed) following blockchain convention.
func (h Hash32) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%x", (&h).Reverse()))
}

// UnmarshalJSON implements json.Unmarshaler.
// Accepts little-endian hex strings (reversed) and stores in big-endian (natural order).
// nolint:revive
func (h *Hash32) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	decoded, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	if len(decoded) != 32 {
		return ErrInvalidHashLength
	}
	// Input is little-endian, reverse to store as big-endian
	for i := 0; i < 32; i++ {
		h[i] = decoded[31-i]
	}
	return nil
}

// IsZero returns true if the hash is all zeros
func (h *Hash32) IsZero() bool {
	return *h == Hash32{}
}
func BytesToHash32(b []byte) (Hash32, error) {
	var h Hash32
	if len(b) != 32 {
		return h, ErrInvalidHashLength
	}

	copy(h[:], b)
	return h, nil
}
func (h *Hash32) Equals(other *Hash32) bool {
	return *h == *other
}

type PrivateKey [32]byte

type PublicKey [33]byte

func (p PublicKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(p[:]))
}

// PublicAddress represents a 20-byte HASH160 address.
type PublicAddress [20]byte

func (pa PublicAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(pa[:]))
}

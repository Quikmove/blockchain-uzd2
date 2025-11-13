package domain

import (
	"encoding/hex"
	"encoding/json"
)

type Hash32 [32]byte

func (h *Hash32) String() string {
	return h.StringBE()
}

func (h *Hash32) StringBE() string {
	reversed := h.Reverse()
	return hex.EncodeToString(reversed[:])
}

func (h *Hash32) StringLE() string {
	return hex.EncodeToString(h[:])
}

func (h *Hash32) Reverse() Hash32 {
	var reversed Hash32
	for i := range 32 {
		reversed[i] = h[31-i]
	}
	return reversed
}

func (h Hash32) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.String())
}

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
	for i := 0; i < 32; i++ {
		h[i] = decoded[31-i]
	}
	return nil
}

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

type PublicAddress [20]byte

func (pa PublicAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(pa[:]))
}

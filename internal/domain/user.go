package domain

import (
	"time"
)

type User struct {
	ID         uint32 `json:"id"`
	Name       string `json:"name"`
	CreatedAt  uint32 `json:"created_at"`
	PublicKey  []byte `json:"public_key"` // Only the public key is stored with the user
	PrivateKey []byte
}

func NewUser(id uint32, name string, publicKey, privateKey []byte) *User {
	return &User{
		ID:         id,
		Name:       name,
		CreatedAt:  uint32(time.Now().Unix()),
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}
}

// Address returns the user's public key as their blockchain address
func (u *User) Address() []byte {
	return u.PublicKey
}

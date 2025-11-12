package domain

import (
	"time"

	"github.com/Quikmove/blockchain-uzd2/internal/crypto"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

type User struct {
	ID            uint32    `json:"id"`
	Name          string    `json:"name"`
	CreatedAt     uint32    `json:"created_at"`
	PublicKey     PublicKey `json:"public_key"`
	PublicAddress PublicAddress
	PrivateKey    PrivateKey
}

func NewUser(id uint32, name string, publicKey PublicKey, privateKey PrivateKey) *User {
	// Generate the public address from the public key
	addressBytes := crypto.GenerateAddress(publicKey[:])
	var publicAddress PublicAddress
	copy(publicAddress[:], addressBytes[:])

	return &User{
		ID:            id,
		Name:          name,
		CreatedAt:     uint32(time.Now().Unix()),
		PublicKey:     publicKey,
		PrivateKey:    privateKey,
		PublicAddress: publicAddress, // Assign the generated address
	}
}

// Address returns the user's pre-calculated HASH160 public address.
func (u *User) Address() PublicAddress {
	return u.PublicAddress
}

// GetPrivateKeyObject converts the raw private key bytes into a secp256k1.PrivateKey object.
func (u *User) GetPrivateKeyObject() *secp256k1.PrivateKey {
	return secp256k1.PrivKeyFromBytes(u.PrivateKey[:])
}

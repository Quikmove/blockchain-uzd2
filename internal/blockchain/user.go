package blockchain

import (
	"math/rand"

	d "github.com/Quikmove/blockchain-uzd2/internal/domain"
)

func GetUserByPublicKey(users []d.User, pk []byte) (d.User, error) {
	if len(pk) == 0 {
		return d.User{}, d.ErrInvalidPublicKey
	}

	for _, u := range users {
		if len(u.PublicKey) != len(pk) {
			continue
		}
		for i, b := range u.PublicKey {
			if b != pk[i] {
				break
			}
			return u, nil
		}
	}
	return d.User{}, d.ErrUserNotFound
}

type UserGeneratorService struct {
	keyGen KeyGenerator
}

func (ugs *UserGeneratorService) GenerateUsers(names []string, n int) []d.User {
	var users []d.User
	id := uint32(1)
	namesLen := len(names)
	usedNames := make(map[string]bool)

	if namesLen == 0 || n <= 0 {
		return []d.User{}
	}
	for i := 0; i < n; i++ {
		name := names[rand.Intn(namesLen)]
		for usedNames[name] {
			name = names[rand.Intn(namesLen)]
		}
		pubKey, privateKey, err := ugs.keyGen.GenerateKeyPair()
		if err != nil {
			panic(err)
		}
		usedNames[name] = true
		user := d.NewUser(
			id,
			name,
			pubKey,
			privateKey,
		)
		id++

		users = append(users, *user)
	}
	return users
}

package blockchain

import (
	"crypto/sha256"
	"math/rand"
	"strconv"
	"time"

	"github.com/Quikmove/blockchain-uzd2/internal/crypto"
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
	keyGen crypto.KeyGenerator
}

func NewUserGeneratorService(keyGen crypto.KeyGenerator) *UserGeneratorService {
	return &UserGeneratorService{keyGen: keyGen}
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
		mneumonicBytes := sha256.Sum256([]byte(strconv.FormatInt(time.Now().UnixNano(), 36)))
		mneumonicString := string(mneumonicBytes[:])
		privateKey, publicKey, err := ugs.keyGen.GenerateKeyPair(mneumonicString)
		if err != nil {
			panic(err)
		}
		usedNames[name] = true
		user := d.NewUser(
			id,
			name,
			publicKey,
			privateKey,
		)
		id++

		users = append(users, *user)
	}
	return users
}

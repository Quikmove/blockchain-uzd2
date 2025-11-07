package blockchain

import (
	"fmt"
	"math/rand"
	"time"
)

type User struct {
	Id        uint32 `json:"id"`
	Name      string `json:"name"`
	CreatedAt uint32 `json:"created_at"`
	PublicKey Hash32 `json:"public_key"`
	// add signing with private key
}

func (u *User) Sign(hash Hash32) ([]byte, error) {
	return []byte("signed-" + u.Name), nil
}

func NewUser(name string) *User {
	id := userCount.Add(1)
	created := uint32(time.Now().Unix())

	data := fmt.Sprintf("%d:%s:%d", id, name, time.Now().UnixNano())

	pk, err := HashString(data, NewArchasHasher())
	if err != nil {
		panic(err)
	}

	return &User{
		Id:        id,
		Name:      name,
		CreatedAt: created,
		PublicKey: pk,
	}
}
func GetUserByPublicKey(users []User, pk Hash32) (User, error) {
	for _, u := range users {
		if u.PublicKey == pk {
			return u, nil
		}
	}
	return User{}, fmt.Errorf("user with public key %x not found", pk)
}
func GenerateUsers(names []string, n int) []User {
	var users []User
	namesLen := len(names)
	if namesLen == 0 || n <= 0 {
		return []User{}
	}
	for i := 0; i < n; i++ {
		user := NewUser(names[rand.Intn(namesLen)])
		users = append(users, *user)
	}
	return users
}

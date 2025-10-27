package blockchain

import (
	"fmt"
	"math/rand"
	"time"
)

type User struct {
	Id        uint32
	Name      string
	CreatedAt uint32
	PublicKey Hash32
}

func NewUser(name string) *User {
	id := userCount.Add(1)
	created := uint32(time.Now().Unix())

	data := fmt.Sprintf("%d:%s:%d", id, name, time.Now().UnixNano())

	pk := HashString(data)

	return &User{
		Id:        id,
		Name:      name,
		CreatedAt: created,
		PublicKey: pk,
	}
}

func GenerateUsers(names []string, n int) []User {
	var Users []User
	namesLen := len(names)
	if namesLen == 0 || n <= 0 {
		return []User{}
	}
	for range n {
		user := NewUser(names[rand.Intn(namesLen)])
		Users = append(Users, *user)
	}
	return Users
}

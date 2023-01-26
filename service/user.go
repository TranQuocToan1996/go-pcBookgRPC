package service

import (
	"bytes"
	"encoding/gob"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UserName string
	HashPw   string
	Role     string
}

func (u *User) IsCorrectPw(rawPw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.HashPw), []byte(rawPw)) == nil
}

func (u *User) Clone() *User {
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(u); err != nil {
		return nil
	}

	clone := &User{}

	if err := gob.NewDecoder(buf).Decode(clone); err != nil {
		return nil
	}

	return clone
}

func NewUser(username, rawPw, role string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(rawPw), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &User{
		UserName: username,
		HashPw:   string(hash),
		Role:     role,
	}, nil
}

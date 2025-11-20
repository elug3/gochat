package model

import (
	"time"

	"github.com/alexedwards/argon2id"
)

type Token struct {
	UserId       int32
	AccessToken  string
	RefreshToken string
	ExpiresIn    time.Duration
	TokenType    string
}

type User struct {
	Id       int32
	Username string
}

type PasswordCredential struct {
	UserId       int32
	Username     string
	PasswordHash string
}

func (pw *PasswordCredential) ValidatePassword(password string) (match bool) {
	match, err := argon2id.ComparePasswordAndHash(password, pw.PasswordHash)
	if err != nil {
		panic(err)
	}
	return match
}

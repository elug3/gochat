package domain

import (
	"net"
	"time"

	"github.com/alexedwards/argon2id"
)

type Session struct {
	SessionId string    `json:"session_id"`
	UserId    int32     `json:"user_id"`
	IP        net.IP    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	RevokedAt time.Time `json:"revoked_at,omitzero"`
}

type Token struct {
	UserId      int32     `json:"user_id"`
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
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

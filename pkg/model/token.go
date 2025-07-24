package model

import (
	"time"
)

type Token struct {
	Id          string     `json:"id"`
	AccessToken string     `json:"access_token"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IssuedAt    time.Time  `json:"issued_at"`
	UserId      int        `json:"user_id"`
}

func (token Token) String() string {
	return token.AccessToken
}

package model

import "time"

type Token struct {
	UserId       int32
	AccessToken  string
	RefreshToken string
	ExpiresIn    time.Duration
	TokenType    string
}

type Credentials struct {
	UserId       int32
	Username     string
	PasswordHash string
	UpdatedAt    time.Time
}

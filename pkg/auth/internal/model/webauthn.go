package model

import (
	"strconv"

	"github.com/go-webauthn/webauthn/webauthn"
)

type WebAuthnUser struct {
	Id          int32
	Name        string
	DisplayName string
	Icon        string
	Credentials []webauthn.Credential
}

func (u *WebAuthnUser) WebAuthnID() []byte {
	return UserHandlerIdBytes(u.Id)
}

func (u *WebAuthnUser) WebAuthnName() string {
	return u.Name
}

func (u *WebAuthnUser) WebAuthnDisplayName() string {
	return u.DisplayName
}

func (u *WebAuthnUser) WebAuthnIcon() string {
	return u.Icon
}

func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}

func UserHandlerIdBytes(userId int32) []byte {
	return []byte(strconv.FormatInt(int64(userId), 10))
}

func UserIdFromUserHandler(userHandle []byte) (int32, error) {
	id, err := strconv.ParseInt(string(userHandle), 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(id), nil
}

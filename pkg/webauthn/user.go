package webauthn

import "github.com/go-webauthn/webauthn/webauthn"

type User struct {
	Id          []byte
	Name        string
	DisplayName string
	Creds       []webauthn.Credential
}

func (u *User) WebAuthnID() []byte {
	return u.Id
}

func (u *User) WebAuthnName() string {
	return u.Name
}

func (u *User) WebAuthnDisplayName() string {
	return u.DisplayName
}

func (u *User) WebAuthnIcon() string {
	return ""
}

func (u *User) WebAuthnCredentials() []webauthn.Credential {
	return u.Creds
}

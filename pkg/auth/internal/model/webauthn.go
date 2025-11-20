package model

import "github.com/go-webauthn/webauthn/webauthn"

type WebAuthnUser struct {
	Id          []byte
	Name        string
	DisplayName string
	Icon        string
	Credentials []webauthn.Credential
}

func (u *WebAuthnUser) WebAuthnID() []byte {
	return u.Id
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

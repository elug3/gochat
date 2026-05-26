package domain

import (
	"strconv"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

// WebAuthnUser represents a user in the context of WebAuthn authentication
type WebAuthnUser struct {
	Id          int32
	Name        string
	DisplayName string
	Icon        string
	Credentials []webauthn.Credential
}

// WebAuthnID implements the webauthn.User interface
func (u *WebAuthnUser) WebAuthnID() []byte {
	return UserHandlerIdBytes(u.Id)
}

// WebAuthnName implements the webauthn.User interface
func (u *WebAuthnUser) WebAuthnName() string {
	return u.Name
}

// WebAuthnDisplayName implements the webauthn.User interface
func (u *WebAuthnUser) WebAuthnDisplayName() string {
	return u.DisplayName
}

// WebAuthnIcon implements the webauthn.User interface
func (u *WebAuthnUser) WebAuthnIcon() string {
	return u.Icon
}

// WebAuthnCredentials implements the webauthn.User interface
func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}

// UserHandlerIdBytes converts a user ID to bytes for WebAuthn user handle
func UserHandlerIdBytes(userId int32) []byte {
	return []byte(strconv.FormatInt(int64(userId), 10))
}

// UserIdFromUserHandler converts a user handle (bytes) back to a user ID
func UserIdFromUserHandler(userHandle []byte) (int32, error) {
	id, err := strconv.ParseInt(string(userHandle), 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(id), nil
}

// Passkey represents a stored WebAuthn credential for a user
type Passkey struct {
	Id         int32     `json:"id"`
	KeyName    string    `json:"key_name"`
	UserId     int32     `json:"user_id"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
}

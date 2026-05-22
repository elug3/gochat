package domain

// User represents an auth domain user entity.
type User struct {
	ID           int32
	Username     string
	PasswordHash string
	Name         string
}

package ports

import "context"

// User represents a minimal auth user domain model used by service ports.
type User struct {
	ID       int32
	Username string
	Name     string
}

// Session represents a minimal auth session domain model used by service ports.
type Session struct {
	ID        string
	UserID    int32
	Audiences []string
	Scopes    []string
}

// UserRepository defines persistence operations required by auth services.
type UserRepository interface {
	CreateUser(
		ctx context.Context,
		username string,
		passwordHash string,
		name string,
	) (*User, error)

	GetUserByUsername(
		ctx context.Context,
		username string,
	) (*User, error)
}

// SessionRepository defines session persistence operations.
type SessionRepository interface {
	CreateSession(
		ctx context.Context,
		userID int32,
		audiences []string,
		scopes []string,
	) (*Session, error)

	GetSessionByID(
		ctx context.Context,
		sessionID string,
	) (*Session, error)
}

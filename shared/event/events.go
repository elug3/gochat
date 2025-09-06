package event

type Event interface {
	Subject() string
}

const (
	// SubjectUserRegistered is the subject for UserRegistered events.
	SubjectUserRegistered = "user.registered"
)

// UserRegistered is published by AuthService after credentials are created for a user.
type UserRegistered struct {
	UserId    int32
	Timestamp int64
}

func (e UserRegistered) Subject() string {
	return SubjectUserRegistered
}

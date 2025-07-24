package model

type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
}

// // TODO: Use optional fields for initialization
// func NewUser(username, password string) (*User, error) {
// 	if len(username) < 2 {
// 	}
// 	if len(password) < 8 {
// 		return nil, errors.New("password must be at least 8 characters long.")
// 	}

// 	return newUser(username, password)
// }

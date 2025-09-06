package model

type User struct {
	Id       int32  `json:"id"`
	Username string `json:"username"`
}

func (u *User) GetId() int32 {
	return u.Id
}

package model

import (
	"time"
)

type Role string

type Profile struct {
	Id       int        `json:"id"`
	Name     string     `json:"name"`
	Birthday *time.Time `json:"birthday,omitempty"`
}

type Group struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Member struct {
	GroupId   int       `json:"group_id"`
	UserId    int       `json:"user_id"`
	Role      Role      `json:"role"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type Message struct {
	Id        string
	SenderId  string
	ConvId    string
	Content   string
	CreatedAt time.Time
}

package model

import (
	"time"

	"github.com/elug3/gochat/pkg/contacts/access"
)

type Profile struct {
	Id       int32      `json:"id"`
	Name     string     `json:"name"`
	Birthday *time.Time `json:"birthday,omitempty"`
	IconUrl  string     `json:"icon_url,omitempty"`
}

type Group struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Member struct {
	GroupId   int         `json:"group_id"`
	UserId    int32       `json:"user_id"`
	Role      access.Role `json:"role"`
	Name      string      `json:"name,omitempty"`
	CreatedAt time.Time   `json:"created_at,omitempty"`
}

// type Message struct {
// 	Id      int       `json:"id"`
// 	ChatId  int       `json:"chat_id"`
// 	Sender  int       `json:"sender_id"`
// 	Content string    `json:"content"`
// 	SentAt  time.Time `json:"sent_at"`
// }

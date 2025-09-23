package model

import "time"

type Message struct {
	ChatId  int       `json:"chat_id,omitempty"`
	Sender  int32     `json:"sender,omitempty"`
	Content string    `json:"content,omitempty"`
	Id      int       `json:"id,omitempty"`
	SentAt  time.Time `json:"sent_at,omitempty"`
}

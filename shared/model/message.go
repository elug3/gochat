package model

import "time"

type Message struct {
	ChatId  int
	Sender  int32
	Content string
	Id      int
	SentAt  time.Time
}

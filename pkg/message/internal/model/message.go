package model

import "time"

type Message struct {
	ChatId  int
	Sender  int
	Content string
	Id      int
	SentAt  time.Time
}

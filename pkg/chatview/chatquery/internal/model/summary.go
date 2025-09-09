package model

type ChatType string

const (
	ChatTypeDirect ChatType = "direct"
	ChatTypeGroup  ChatType = "group"
)

type ChatSummary struct {
	ChatId        int      `json:"chatId,string"`
	Type          ChatType `json:"type"`
	Name          string   `json:"name"`
	LastMessage   string   `json:"last_message"`
	LastMessageAt int64    `json:"last_message_at,string"`
	LastSenderId  int      `json:"last_sender_id"`
	UnreadCount   int      `json:"unread_count,string"`
	Participants  []int    `json:"participants"`
}

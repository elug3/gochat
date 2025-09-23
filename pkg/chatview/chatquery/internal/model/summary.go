package model

type ChatType string

const (
	ChatTypeDirect ChatType = "direct"
	ChatTypeGroup  ChatType = "group"
)

type ChatSummary struct {
	Id            int      `json:"id"`
	Type          ChatType `json:"type"`
	Name          string   `json:"name"`
	LastMessage   string   `json:"last_message"`
	LastMessageAt int64    `json:"last_message_at"`
	LastSenderId  int      `json:"last_sender_id"`
	UnreadCount   int      `json:"unread_count"`
	Pinned        bool     `json:"pinned"`
	Muted         bool     `json:"muted"`
}

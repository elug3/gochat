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
	LastMessage   string   `json:"lastMessage"`
	LastMessageAt int64    `json:"lastMessageAt,string"`
	LastSenderId  int      `json:"lastSenderId"`
	UnreadCount   int      `json:"unreadCount,string"`
	Pinned        bool     `json:"pinned"`
	Muted         bool     `json:"muted"`
}

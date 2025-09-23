package store

import (
	"context"
)

type ChatStore interface {
	CreateGroupChat(ctx context.Context, chatId int, name string, timestamp int64) error
	// UpdateLastMessage updates last message of the chat and increments the chat sequence.
	UpdateLastMessage(ctx context.Context, chatId int, message string, timestamp int64) error
	// AddChatToUser adds chat to user's chat list.
	AddChatToUser(ctx context.Context, userId int32, chatId int, timestamp int64) error

	UpdateLastReadSeq(ctx context.Context, userId int32, chatId int) error
}

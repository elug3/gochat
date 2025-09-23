package store

import (
	"context"

	"github.com/elug3/gochat/pkg/message/internal/model"
)

type MessageOptions struct {
	Limit  int
	Offset int
}

type MessageStore interface {
	WithContext(ctx context.Context) (context.Context, context.CancelFunc)
	CreateMessage(ctx context.Context, chatId int, userId int32, content string) (*model.Message, error)
	ListMessages(ctx context.Context, chatId int, options *MessageOptions) ([]model.Message, error)
	DeleteMessage(ctx context.Context, messageId int) error
}

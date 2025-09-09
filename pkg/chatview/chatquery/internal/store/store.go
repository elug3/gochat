package store

import (
	"context"

	"github.com/elug3/gochat/pkg/chatview/chatquery/internal/model"
)

type ChatViewStore interface {
	ListByUserId(ctx context.Context, userId int32) ([]model.ChatSummary, error)
}

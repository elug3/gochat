package chatquery

import (
	"context"

	"github.com/elug3/gochat/pkg/chatview/chatquery/internal/model"
	"github.com/elug3/gochat/pkg/chatview/chatquery/internal/store"
	redisstore "github.com/elug3/gochat/pkg/chatview/chatquery/internal/store/redis-store"
)

type ChatViewService struct {
	store store.ChatViewStore
}

func NewChatViewService(opts *Options) (*ChatViewService, error) {
	store, err := redisstore.NewChatViewStore(opts.DatabaseURL)
	if err != nil {
		return nil, err
	}
	return &ChatViewService{
		store: store,
	}, nil
}

func (s *ChatViewService) ListChats(ctx context.Context, userId int32) ([]model.ChatSummary, error) {
	return s.store.ListByUserId(ctx, userId)
}

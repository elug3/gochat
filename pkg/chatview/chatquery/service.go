package chatquery

import (
	"context"
	"fmt"

	"github.com/elug3/gochat/pkg/chatview/chatquery/internal/model"
	"github.com/elug3/gochat/pkg/chatview/chatquery/internal/store"
)

type ChatViewService struct {
	store store.ChatViewStore
}

type ServiceDeps struct {
	Store store.ChatViewStore
}

func NewChatViewService(deps ServiceDeps) (*ChatViewService, error) {
	if deps.Store == nil {
		return nil, fmt.Errorf("chat view store is required")
	}
	return &ChatViewService{
		store: deps.Store,
	}, nil
}

func (s *ChatViewService) ListChats(ctx context.Context, userId int32) ([]model.ChatSummary, error) {
	return s.store.ListByUserId(ctx, userId)
}

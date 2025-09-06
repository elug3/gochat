package service

import (
	"context"

	"github.com/elug3/gochat/pkg/events"
	"github.com/elug3/gochat/pkg/model"
	"github.com/elug3/gochat/pkg/store"
)

type ChatProjectionService struct {
	store store.ChatStore
}

// event handlers for chat projection service
func NewChatProjectionService(chatStore store.ChatStore) *ChatProjectionService {
	return &ChatProjectionService{
		store: chatStore,
	}
}

func (s *ChatProjectionService) OnGroupCreated(ctx context.Context, ev *events.GroupCreated) error {
	return s.store.CreateGroupChat(ctx, ev.GroupId, ev.GroupName, ev.TimeStamp)
}

func (s *ChatProjectionService) OnGroupUpdated(ctx context.Context, ev *events.GroupUpdated) error {
	return nil
}

func (s *ChatProjectionService) OnMessageSent(ctx context.Context, ev *events.MessageSent) error {
	return s.store.UpdateLastMessage(ctx, ev.ChatId, ev.Content, ev.Timestamp)
}

func (s *ChatProjectionService) OnMemberJoined(ctx context.Context, ev *events.MemberJoined) error {
	return s.store.AddChatToUser(ctx, ev.UserId, ev.GroupId, ev.TimeStamp)
}

// query methods for chat projection service
func (s *ChatProjectionService) ListChats() ([]model.ChatSummary, error) {
	return nil, nil
}

func (s *ChatProjectionService) GetChat() (*model.ChatSummary, error) {
	return nil, nil
}

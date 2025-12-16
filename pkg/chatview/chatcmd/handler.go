package chatcmd

import (
	"context"

	redisstore "github.com/elug3/gochat/pkg/chatview/chatcmd/internal/store/redis-store"
	"github.com/elug3/gochat/shared/events"
)

type ChatCommandHandler struct {
	store *redisstore.ChatStore
}

func NewChatCommandHandler(opts *Options) (*ChatCommandHandler, error) {
	store, err := redisstore.NewChatStore(opts.DatabaseUrl)
	if err != nil {
		return nil, err
	}
	return &ChatCommandHandler{store: store}, nil
}

func (s *ChatCommandHandler) OnGroupCreated(ctx context.Context, ev *events.GroupCreated) error {
	return s.store.CreateGroupChat(ctx, ev.GroupId, ev.GroupName, ev.TimeStamp)
}

func (s *ChatCommandHandler) OnGroupDeleted(ctx context.Context, ev *events.GroupDeleted) error {
	return s.store.DeleteChatMeta(ctx, ev.GroupId)
}

func (s *ChatCommandHandler) OnMemberJoined(ctx context.Context, ev *events.MemberJoined) error {
	return s.store.AddChatToUser(ctx, ev.UserId, ev.GroupId, ev.TimeStamp)
}

func (s *ChatCommandHandler) OnMemberLeft(ctx context.Context, ev *events.MemberLeft) error {
	// return s.store.RemoveChatFromUser(ctx, ev.UserId, ev.GroupId, ev.TimeStamp)
	return nil
}

func (s *ChatCommandHandler) OnMessageSent(ctx context.Context, ev *events.MessageSent) error {
	return s.store.UpdateLastMessage(ctx, ev.ChatId, ev.Content, ev.TimeStamp)
}

func (s *ChatCommandHandler) OnMessageRead(ctx context.Context, ev *events.MessageRead) error {
	return s.store.UpdateLastReadSeq(ctx, ev.ChatId, ev.UserId)
}

func (s *ChatCommandHandler) OnContactsReset(ctx context.Context, ev *events.ContactsReset) error {
	return s.store.Reset(ctx)
}

package chatcmd

import (
	"context"
	"encoding/json"

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

func (s *ChatCommandHandler) OnGroupCreated(ctx context.Context, subject string, data []byte) error {
	var ev events.GroupCreated
	err := json.Unmarshal(data, &ev)
	if err != nil {
		return err
	}
	return s.store.CreateGroupChat(ctx, ev.GroupId, ev.GroupName, ev.Timestamp)
}

func (s *ChatCommandHandler) OnGroupDeleted(ctx context.Context, subject string, data []byte) error {
	var ev events.GroupDeleted
	err := json.Unmarshal(data, &ev)
	if err != nil {
		return err
	}
	return s.store.DeleteChatMeta(ctx, ev.GroupId)
}

func (s *ChatCommandHandler) OnMemberJoined(ctx context.Context, subject string, data []byte) error {
	var ev events.MemberJoined
	err := json.Unmarshal(data, &ev)
	if err != nil {
		return err
	}
	return s.store.AddChatToUser(ctx, ev.UserId, ev.GroupId, ev.Timestamp)
}

func (s *ChatCommandHandler) OnMemberLeft(ctx context.Context, subject string, data []byte) error {
	var ev events.MemberLeft
	err := json.Unmarshal(data, &ev)
	if err != nil {
		return err
	}
	return nil
}

func (s *ChatCommandHandler) OnMessageSent(ctx context.Context, subject string, data []byte) error {
	var ev events.MessageSent
	err := json.Unmarshal(data, &ev)
	if err != nil {
		return err
	}
	return s.store.UpdateLastMessage(ctx, ev.ChatId, ev.Content, ev.Timestamp)
}

func (s *ChatCommandHandler) OnMessageRead(ctx context.Context, subject string, data []byte) error {
	var ev events.MessageRead
	err := json.Unmarshal(data, &ev)
	if err != nil {
		return err
	}
	return s.store.UpdateLastReadSeq(ctx, ev.ChatId, ev.UserId)
}

func (s *ChatCommandHandler) OnContactsReset(ctx context.Context, subject string, data []byte) error {
	var ev events.ContactsReset
	err := json.Unmarshal(data, &ev)
	if err != nil {
		return err
	}
	return s.store.Reset(ctx)
}

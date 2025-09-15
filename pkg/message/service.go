package message

import (
	"context"
	"fmt"

	"github.com/elug3/gochat/api/httpclient"
	"github.com/elug3/gochat/pkg/message/internal/errs"
	"github.com/elug3/gochat/pkg/message/internal/model"
	"github.com/elug3/gochat/pkg/message/internal/store"
	"github.com/elug3/gochat/pkg/message/internal/store/sqlite3"
	"github.com/elug3/gochat/shared/events"
)

type MessageService struct {
	store    store.MessageStore
	contacts *httpclient.ContactsClient
	pub      *events.Publisher
}

func NewMessageService(opts *Options) (*MessageService, error) {

	store, err := sqlite3.NewMessageStore(opts.SaveDir, opts.NoSave)
	if err != nil {
		return nil, fmt.Errorf("cannot create message store: %w", err)
	}
	pub, err := events.NewPublisher(opts.NatsUrl)
	if err != nil {
		return nil, fmt.Errorf("cannot create event publisher: %w", err)
	}

	contactsClient := httpclient.NewContactsClient(
		httpclient.WithBaseUrl(opts.ContactsServerUrl),
	)
	s := MessageService{
		store:    store,
		pub:      pub,
		contacts: contactsClient,
	}
	return &s, nil
}

func (s *MessageService) Send(ctx context.Context, userId, chatId int, content string) (*model.Message, error) {
	var msg *model.Message
	var err error

	ctx, cancel := s.store.WithContext(ctx)
	defer cancel()

	// positive chatId sends to group, negative is direct message to user
	if chatId > 0 {
		msg, err = s.sendToGroup(ctx, userId, chatId, content)
		if err != nil {
			return nil, fmt.Errorf("cannot send to group '%d': %w", chatId, err)
		}
	} else if chatId < 0 {
		msg, err = s.sendToUser(ctx, userId, chatId, content)
		if err != nil {
			return nil, fmt.Errorf("cannot send to user '%d': %w", userId, err)
		}
	} else {
		return nil, fmt.Errorf("invalid chatId: %d", chatId)
	}

	s.pub.Publish(events.MessageSent{
		ChatId:    chatId,
		SenderId:  userId,
		Content:   content,
		TimeStamp: msg.SentAt.Unix(),
	})
	return msg, nil
}

// ListMessages gets messages for a chat without permission checks
func (s *MessageService) ListMessages(ctx context.Context, chatId int) ([]model.Message, error) {
	ctx, cancel := s.store.WithContext(ctx)
	defer cancel()

	messages, err := s.store.ListMessages(ctx, chatId, &store.MessageOptions{
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot list messages: %w", err)
	}
	return messages, nil
}

// ListUserChatMessages gets messages for a chat with permission checks
func (s *MessageService) ListUserChatMessages(ctx context.Context, chatId, userId int) ([]model.Message, error) {
	ctx, cancel := s.store.WithContext(ctx)
	defer cancel()

	resp, err := s.contacts.Can(ctx, httpclient.AccessRequestParams{
		ChatId: chatId,
		UserId: userId,
		Action: "read", // TODO: use constant
	})
	if err != nil {
		return nil, err
	}
	if !resp.Can {
		return nil, fmt.Errorf("cannot list messages: %w", errs.PermissionDenied)
	}

	messages, err := s.store.ListMessages(ctx, chatId, &store.MessageOptions{
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot list messages: %w", err)
	}
	return messages, nil
}

func (s *MessageService) sendToGroup(ctx context.Context, userId, chatId int, content string) (*model.Message, error) {
	ctx, cancel := s.store.WithContext(ctx)
	defer cancel()

	resp, err := s.contacts.Can(ctx, httpclient.AccessRequestParams{
		ChatId: chatId,
		UserId: userId,
		Action: "send", // TODO: use constant
	})
	if err != nil {
		return nil, err
	}
	if !resp.Can {
		return nil, fmt.Errorf("cannot send message: %w", errs.PermissionDenied)
	}

	msg, err := s.store.CreateMessage(ctx, chatId, userId, content)
	if err != nil {
		return nil, fmt.Errorf("cannot create message: %w", err)
	}

	return msg, nil
}

// TODO: Implement sending to user
func (s *MessageService) sendToUser(ctx context.Context, chatId, userId int, content string) (*model.Message, error) {
	return nil, nil

}

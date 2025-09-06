package service

import (
	"context"
	"fmt"

	"github.com/elug3/gochat/internal/message/model"
	"github.com/elug3/gochat/pkg/events"
	"github.com/elug3/gochat/pkg/store"
)

type MessageService struct {
	Contacts *ContactsService
	store    store.MessageStore
	ep       *events.EventPublisher
}

func NewMessageService(
	messageStore store.MessageStore,
	contactsService *ContactsService,
	ep *events.EventPublisher,
) (*MessageService, error) {
	s := MessageService{
		store:    messageStore,
		Contacts: contactsService,
		ep:       ep,
	}
	return &s, nil
}

func (s *MessageService) Send(ctx context.Context, userId, chatId int, content string) (*model.Message, error) {
	var msg *model.Message
	var err error

	ctx, cancel := s.store.WithContext(ctx)
	defer cancel()

	if chatId > 0 {
		msg, err = s.sendToGroup(ctx, userId, chatId, content)
		if err != nil {
			return nil, fmt.Errorf("cannot send to group '%d': %w", chatId, err)
		}
	} else {
		msg, err = s.sendToUser(ctx, userId, chatId, content)
		if err != nil {
			return nil, fmt.Errorf("cannot send to user '%d': %w", userId, err)
		}
	}
	if err = s.ep.Publish(events.MessageSent{
		ChatId:    chatId,
		SenderId:  userId,
		Content:   content,
		Timestamp: msg.SentAt.Unix(),
	}); err != nil {
		return nil, fmt.Errorf("Publish: %w", err)
	}
	return msg, nil
}

func (s *MessageService) ListMessages(ctx context.Context, groupId, userId int) ([]model.Message, error) {
	ctx, cancel := s.store.WithContext(ctx)
	defer cancel()

	if ok, err := s.Contacts.CanRead(groupId, userId); !ok {
		if err != nil {
			return nil, fmt.Errorf("CanRead: %w", err)
		}
		return nil, &store.Error{
			Kind:    store.KindMessage,
			Message: "permission denied",
		}
	}
	messages, err := s.store.ListMessages(ctx, groupId, &store.MessageOptions{
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

	if ok, err := s.Contacts.CanSend(chatId, userId); !ok {
		if err != nil {
			return nil, fmt.Errorf("CanSend: %w", err)
		}
		return nil, &store.Error{
			Kind:    store.KindMessage,
			Message: "permission denied",
		}
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

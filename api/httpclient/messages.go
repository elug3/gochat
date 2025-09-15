package httpclient

import (
	"context"
	"fmt"

	"github.com/elug3/gochat/shared/model"
)

type MessagesClient struct {
	options []RequestOption
}

func NewMessagesClient(opts ...RequestOption) *MessagesClient {
	return &MessagesClient{
		options: opts,
	}
}

func (c *MessagesClient) List(ctx context.Context, chatId int, opts ...RequestOption) ([]model.Message, error) {
	opts = append(c.options, opts...)

	path := fmt.Sprintf("/chats/%d/messages", chatId)
	cfg, err := NewRequestConfig("GET", path, nil, opts...)
	if err != nil {
		return nil, err
	}

	var result []model.Message
	if err = cfg.Do(ctx, &result); err != nil {
		return nil, err
	}

	return result, nil
}

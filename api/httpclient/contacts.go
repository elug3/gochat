package httpclient

import (
	"context"
	"fmt"

	"github.com/elug3/gochat/shared/model"
	"github.com/google/go-querystring/query"
)

type ContactsClient struct {
	options []RequestOption
	Groups  *GroupsClient
}

func NewContactsClient(opts ...RequestOption) *ContactsClient {
	return &ContactsClient{
		options: opts,
		Groups:  NewGroupsClient(opts...),
	}
}

type AccessRequestParams struct {
	UserId   int32  `url:"user_id"`
	ChatId   int    `url:"chat_id"`
	TargetId int32  `url:"target_id"`
	Action   string `url:"action"`
}

type AccessResponse struct {
	Can    bool   `json:"can"`
	Action string `json:"action"`
	Error  string `json:"error"`
}

func (c *ContactsClient) Can(ctx context.Context, params AccessRequestParams, opts ...RequestOption) (*AccessResponse, error) {
	opts = append(c.options, opts...)

	q, err := query.Values(params)
	if err != nil {
		return nil, err
	}

	cfg, err := NewRequestConfig("GET", "/can?"+q.Encode(), nil, opts...)
	if err != nil {
		return nil, err
	}

	var result AccessResponse
	if err = cfg.Do(ctx, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

type GroupsClient struct {
	opts []RequestOption
}

func NewGroupsClient(opts ...RequestOption) *GroupsClient {
	return &GroupsClient{
		opts: opts,
	}
}

func (c *GroupsClient) List(ctx context.Context, opts ...RequestOption) ([]model.Group, error) {
	opts = append(c.opts, opts...)

	cfg, err := NewRequestConfig("GET", "/groups", nil, opts...)
	if err != nil {
		return nil, err
	}

	groups := make([]model.Group, 0)
	if err = cfg.Do(ctx, &groups); err != nil {
		return nil, err
	}

	return groups, nil
}

func (c GroupsClient) ListByUser(ctxx context.Context, userId int32, opts ...RequestOption) ([]model.Group, error) {
	opts = append(c.opts, opts...)

	path := fmt.Sprintf("/user/%d/groups", userId)
	cfg, err := NewRequestConfig("GET", path, nil, opts...)
	if err != nil {
		return nil, err
	}

	groups := make([]model.Group, 0)
	if err = cfg.Do(ctxx, &groups); err != nil {
		return nil, err
	}

	return groups, nil
}

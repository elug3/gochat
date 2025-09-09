package redisstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/elug3/gochat/pkg/chatview/chatquery/internal/model"
	"github.com/redis/go-redis/v9"
	"github.com/tidwall/gjson"
)

type ChatViewStore struct {
	rdb *redis.Client
}

func NewChatViewStore() (*ChatViewStore, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	err := rdb.Ping(context.TODO()).Err()
	if err != nil {
		return nil, err
	}
	return &ChatViewStore{
		rdb: rdb,
	}, nil
}

func (store *ChatViewStore) ListByUserId(ctx context.Context, userId int32) ([]model.ChatSummary, error) {
	s, err := listUserChatSummariesScript.Run(ctx, store.rdb, []string{strconv.FormatInt(int64(userId), 10)}).Text()
	if err != nil {
		return nil, err
	}

	if !gjson.Valid(s) {
		return nil, fmt.Errorf("invalid json")
	}
	jr := gjson.Parse(s)
	if errStr := jr.Get("err").String(); errStr != "" {
		return nil, errors.New(errStr)
	}
	var chats []model.ChatSummary
	for _, v := range jr.Get("chats").Array() {
		var chat model.ChatSummary
		err := json.Unmarshal([]byte(v.Raw), &chat)
		if err != nil {
			return nil, err
		}

		chats = append(chats, chat)
	}

	return chats, nil
}

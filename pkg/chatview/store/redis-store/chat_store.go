package redisstore

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/elug3/gochat/pkg/config"
	"github.com/elug3/gochat/pkg/model"
	"github.com/elug3/gochat/pkg/store"
	"github.com/redis/go-redis/v9"
)

func init() {
	store.RegisterDriver("redis", func(cfg *config.Config) (store.ChatStore, error) {
		return NewChatStore(cfg)

	})
}

type ChatStore struct {
	rdb *redis.Client
}

func NewChatStore(cfg *config.Config) (*ChatStore, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()
	// TODO: unused config
	addr := "localhost:6379"
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	rdb.Ping(ctx).Err()
	err := rdb.Ping(context.Background()).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	store := ChatStore{
		rdb: rdb,
	}

	return &store, nil
}

func (store *ChatStore) CreateGroupChat(ctx context.Context, chatId int, groupName string, timestamp int64) error {
	keys := []string{
		strconv.Itoa(chatId),
	}
	args := []string{
		groupName,
		strconv.FormatInt(timestamp, 10),
	}
	s, err := createGroupMetaScript.Run(ctx, store.rdb, keys, args).Text()
	if err != nil {
		return err
	}
	res := struct {
		Added int
		Err   string
	}{}
	if err = json.Unmarshal([]byte(s), &res); err != nil {
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}
	if res.Err != "" {
		return fmt.Errorf("failed to create group chat: %s", res.Err)
	}
	return nil
}

func (store *ChatStore) AddChatToUser(ctx context.Context, userId int, chatId int, timestamp int64) error {
	keys := []string{
		strconv.Itoa(userId),
		strconv.Itoa(chatId),
	}
	s, err := addChatToUserScript.Run(ctx, store.rdb, keys).Text()
	if err != nil {
		return err
	}
	res := struct {
		Added int
		Err   string
	}{}
	if err = json.Unmarshal([]byte(s), &res); err != nil {
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}
	if res.Err != "" {
		return fmt.Errorf("failed to add chat to user: %s", res.Err)
	}
	return nil
}

func (store *ChatStore) UpdateLastMessage(ctx context.Context, chatId int, message string, timestamp int64) error {
	keys := []string{
		strconv.Itoa(chatId),
	}
	args := []string{
		message,
		strconv.FormatInt(timestamp, 10),
	}
	s, err := updateLastMessageScript.Run(ctx, store.rdb, keys, args).Text()
	if err != nil {
		return err
	}
	res := struct {
		Seq int
		Err string
	}{}
	if err = json.Unmarshal([]byte(s), &res); err != nil {
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}
	if res.Err != "" {
		return fmt.Errorf("failed to update last message: %s", res.Err)
	}
	return nil
}

func (store *ChatStore) ListChatSummaries(ctx context.Context, userId int) ([]model.ChatSummary, error) {
	s, err := listUserChatSummariesScript.Run(ctx, store.rdb, []string{strconv.Itoa(userId)}).Text()
	if err != nil {
		return nil, err
	}
	res := struct {
		Chats []model.ChatSummary
		Err   string
	}{}
	if err = json.Unmarshal([]byte(s), &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}
	if res.Err != "" {
		return nil, fmt.Errorf("failed to list chat summaries: %s", res.Err)
	}
	return res.Chats, nil
}

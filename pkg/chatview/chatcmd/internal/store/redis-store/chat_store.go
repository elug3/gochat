package redisstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type ChatStore struct {
	rdb *redis.Client
}

func NewChatStore(redisUrl string) (*ChatStore, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()
	rdb := redis.NewClient(&redis.Options{
		Addr: redisUrl,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
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

// DeleteChatMeta deletes chat metadata and all associated user chat references.
func (store *ChatStore) DeleteChatMeta(ctx context.Context, chatId int) error {
	keys := []string{
		strconv.Itoa(chatId),
	}
	s, err := deleteChatMetaScript.Run(ctx, store.rdb, keys).Text()
	if err != nil {
		return err
	}
	res := struct {
		Deleted int
		Err     string
	}{}
	if err = json.Unmarshal([]byte(s), &res); err != nil {
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}
	if res.Err != "" {
		return fmt.Errorf("failed to delete chat meta: %s", res.Err)
	}
	return nil
}

// DeleteChatHistory deletes chat history in user chat list.
func (store *ChatStore) DeleteChatHistory(ctx context.Context, chatId int, userId int32) error {
	keys := []string{
		strconv.FormatInt(int64(userId), 10),
		strconv.Itoa(chatId),
	}
	s, err := deleteChatHistoryScript.Run(ctx, store.rdb, keys).Text()
	if err != nil {
		return err
	}
	res := struct {
		Removed int
		Err     string
	}{}
	if err = json.Unmarshal([]byte(s), &res); err != nil {
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}
	if res.Err != "" {
		return fmt.Errorf("failed to delete chat history: %s", res.Err)
	}
	return nil
}

func (store *ChatStore) AddChatToUser(ctx context.Context, userId int32, chatId int, timestamp int64) error {
	keys := []string{
		strconv.FormatInt(int64(userId), 10),
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

func (store *ChatStore) UpdateLastReadSeq(ctx context.Context, chatId int, userId int32) error {
	keys := []string{
		strconv.FormatInt(int64(userId), 10), // local userId = KEYS[1]
		strconv.Itoa(chatId),                 // local chatId = KEYS[2]
	}
	fmt.Println("userId:", strconv.FormatInt(int64(userId), 10))

	result, err := updateLastReadSeqScript.Run(ctx, store.rdb, keys).Text()
	if err != nil {
		return fmt.Errorf("script error: %w", err)
	}

	var res struct {
		Err string `json:"err"`
	}
	if err = json.Unmarshal([]byte(result), &res); err != nil {
		return err
	}
	if res.Err != "" {
		return errors.New(res.Err)
	}

	return nil
}

// Reset clears all data in the chat store.
func (store *ChatStore) Reset(ctx context.Context) error {
	if err := store.rdb.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("failed to reset store: %w", err)
	}
	return nil
}

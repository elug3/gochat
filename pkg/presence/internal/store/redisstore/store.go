package redisstore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elug3/gochat/pkg/presence/internal/errs"
	"github.com/elug3/gochat/pkg/presence/internal/model"
	"github.com/redis/go-redis/v9"
)

type PresenceStore struct {
	rdb *redis.Client
}

func NewPresenceStore(redisAddr string) (*PresenceStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return &PresenceStore{
		rdb: client,
	}, nil
}

func (store *PresenceStore) SetUserPresence(ctx context.Context, userId int32, state string) error {
	presence := model.Presence{
		UserId:   userId,
		State:    state,
		LastSeen: time.Now().Unix(),
	}

	payload, err := json.Marshal(presence)
	if err != nil {
		return err
	}

	return store.rdb.Set(ctx, formatPresenceKey(userId), payload, 0).Err()
}

func (store *PresenceStore) GetUserPresence(ctx context.Context, userId int32) (*model.Presence, error) {
	result, err := store.rdb.Get(ctx, formatPresenceKey(userId)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	presence := &model.Presence{}
	if err := json.Unmarshal([]byte(result), presence); err != nil {
		return nil, err
	}

	return presence, nil
}

func formatPresenceKey(userId int32) string {
	return fmt.Sprintf("presence:%d", userId)
}

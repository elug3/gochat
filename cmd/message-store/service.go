package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Message struct {
	Id          int64     `json:"id,omitempty"`
	RoomId      string    `json:"room_id,omitempty"`
	SenderId    string    `json:"sender_id,omitempty"`
	SentAt      time.Time `json:"sent_at,omitempty"`
	IsDeleted   bool      `json:"is_deleted,omitempty"`
	MessageType string    `json:"message_type,omitempty"`
	Content     string    `json:"content,omitempty"`
}

type MessageService struct {
	pool *pgxpool.Pool
}

func NewMessageService() *MessageService {
	pool, _ := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}
	return &MessageService{pool}
}

func (s *MessageService) GetMessages(ctx context.Context, roomId string) ([]Message, error) {
	rows, _ := s.pool.Query(ctx, `
	SELECT (id, room_id, sender_id, sent_at, is_deleted, message_type, content)
	FROM messages
	WHERE room_id = ($1)
	ORDER BY sent_at DESC
	LIMIT 50;`, roomId)
	messages, err := pgx.CollectRows(rows, pgx.RowTo[Message])
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (s *MessageService) PostMessage(ctx context.Context, message *Message) error {
	tag, err := s.pool.Exec(ctx, `
	INSERT INTO messages (room_id, sender_id, content)
	VALUES ($1, $2, $3);`, message.RoomId, message.SenderId, message.Content)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("no message were inserted")
	}
	return nil
}

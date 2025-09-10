package sqlite3

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/elug3/gochat/internal/message/internal/model"
	"github.com/elug3/gochat/internal/message/internal/store"
	_ "github.com/mattn/go-sqlite3"
)

type MessageStore struct {
	db *sql.DB
}

func NewMessageStore(saveDir string, noSave bool) (*MessageStore, error) {
	db, err := openDB(saveDir, noSave)
	if err != nil {
		return nil, fmt.Errorf("cannot open database: %w", err)
	}
	if err = initDB(db); err != nil {
		return nil, fmt.Errorf("cannot initialize database: %w", err)
	}
	store := MessageStore{db: db}
	return &store, nil
}

type TxKey struct{}

func (store *MessageStore) WithContext(ctx context.Context) (context.Context, context.CancelFunc) {
	tx, _ := store.db.BeginTx(ctx, &sql.TxOptions{})

	ctx, cancel := context.WithCancel(ctx)
	ctx = context.WithValue(ctx, TxKey{}, tx)
	return ctx, cancel
}

func GetTx(ctx context.Context) (*sql.Tx, error) {
	tx, ok := ctx.Value(TxKey{}).(*sql.Tx)
	if !ok {
		return nil, fmt.Errorf("no transaction found in context")
	}
	return tx, nil
}

func (s *MessageStore) CreateMessage(ctx context.Context, chatId, userId int, content string) (*model.Message, error) {
	tx, err := GetTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var msg model.Message
	err = tx.QueryRow(`
	INSERT INTO message (chat_id, sender, content)
	VALUES (?, ?, ?)
	RETURNING id, chat_id, sender, content, sent_at;
	`, chatId, userId, content).Scan(&msg.Id, &msg.ChatId, &msg.Sender, &msg.Content, &msg.SentAt)
	if err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return &msg, nil
}

func (s *MessageStore) ListMessages(ctx context.Context, chatId int, options *store.MessageOptions) ([]model.Message, error) {
	tx, err := GetTx(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(`
	SELECT id, chat_id, sender, content, sent_at
	FROM message
	WHERE chat_id = ?
	LIMIT ? OFFSET ?;
	`, chatId, options.Limit, options.Offset)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	messages := make([]model.Message, 0)
	for rows.Next() {
		var msg model.Message
		if err := rows.Scan(&msg.Id, &msg.ChatId, &msg.Sender, &msg.Content, &msg.SentAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return messages, nil
}

func (s *MessageStore) DeleteMessage(ctx context.Context, messageId int) error {
	tx, err := GetTx(ctx)
	if err != nil {
		return err
	}
	result, err := tx.Exec(`DELETE FROM message WHERE id = ?`, messageId)
	if err != nil {
		return fmt.Errorf("delete message: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no message found with id: %d", messageId)
	}
	return nil
}

func openDB(saveDir string, noSave bool) (*sql.DB, error) {
	if noSave {
		return sql.Open("sqlite3", ":memory:")
	}
	path := saveDir + "/messages.db"
	return sql.Open("sqlite3", path)
}

func initDB(db *sql.DB) error {
	var err error
	if _, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS message (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id INTEGER NOT NULL,
		sender INTEGER NOT NULL,
		sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		content TEXT NOT NULL
		);`); err != nil {
		return fmt.Errorf("create table message: %w", err)
	}

	return nil
}

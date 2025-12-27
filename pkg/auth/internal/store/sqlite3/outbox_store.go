package sqlite3

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/elug3/gochat/pkg/auth/internal/model"
)

type OutboxStore struct {
	db *sql.DB
}

func NewOutboxStore(db *sql.DB) (*OutboxStore, error) {
	if db == nil {
		return nil, fmt.Errorf("db is required")
	}
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `
	PRAGMA journal_mode = WAL;
	PRAGMA busy_timeout = 5000;
	PRAGMA synchronous = NORMAL;
	`); err != nil {
		return nil, fmt.Errorf("failed to set sqlite pragmas: %w", err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err = initOutboxEventTable(ctx, tx); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}
	return &OutboxStore{db: db}, nil
}

func (store *OutboxStore) Insert(ctx context.Context, tx *sql.Tx, row *model.OutboxRecord) error {
	if row == nil {
		return fmt.Errorf("outbox record is required")
	}
	if row.Id == "" {
		return fmt.Errorf("outbox record id is required")
	}
	if row.Subject == "" {
		return fmt.Errorf("outbox record subject is required")
	}
	createdAt := row.CreatedAt
	if createdAt == 0 {
		createdAt = time.Now().Unix()
	}
	availableAt := row.AvailableAt
	if availableAt == 0 {
		availableAt = createdAt
	}
	attempts := row.Attempts
	if attempts < 0 {
		attempts = 0
	}
	_, err := store.execContext(ctx, tx, `
	INSERT INTO outbox_events (
		id, subject, payload, created_at, available_at, attempts
	) VALUES (?, ?, ?, ?, ?, ?);
	`, row.Id, row.Subject, row.Payload, createdAt, availableAt, attempts)
	if err != nil {
		return fmt.Errorf("failed to insert outbox event: %w", err)
	}
	return nil
}

func (store *OutboxStore) ClaimBatch(
	ctx context.Context,
	workerId string,
	limit int,
	leaseSeconds int,
) ([]model.OutboxEvent, error) {
	if workerId == "" {
		return nil, fmt.Errorf("workerId is required")
	}
	if limit <= 0 {
		return []model.OutboxEvent{}, nil
	}
	if leaseSeconds <= 0 {
		leaseSeconds = 1
	}
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	nowExpr := "CAST(strftime('%s', 'now') AS INTEGER)"
	claimSQL := fmt.Sprintf(`
	UPDATE outbox_events
	SET status = ?,
		locked_by = ?,
		locked_until = %s + ?,
		attempts = attempts + 1
	WHERE id IN (
		SELECT id
		FROM outbox_events
		WHERE
		(
			status = ?
			AND available_at <= %s
		)
		OR
		(
			status = ?
			AND locked_until IS NOT NULL
			AND locked_until <= %s
		)
	ORDER BY created_at
	LIMIT ?
	);`, nowExpr, nowExpr, nowExpr)

	if _, err = tx.ExecContext(
		ctx,
		claimSQL,
		model.OutboxStatusProcessing,
		workerId,
		leaseSeconds,
		model.OutboxStatusNew,
		model.OutboxStatusProcessing,
		limit,
	); err != nil {
		return nil, fmt.Errorf("failed to set processing status: %w", err)
	}

	selectSQL := fmt.Sprintf(`
	SELECT id, subject, payload,
		status, created_at, available_at,
		COALESCE(locked_by, ''), COALESCE(locked_until, 0),
		attempts, COALESCE(last_error, '')
	FROM outbox_events
	WHERE status = ?
		AND locked_by = ?
		AND locked_until > %s
	ORDER BY created_at
	LIMIT ?;
	`, nowExpr)

	rows, err := tx.QueryContext(ctx, selectSQL, model.OutboxStatusProcessing, workerId, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get claims: %w", err)
	}
	defer rows.Close()

	events := make([]model.OutboxEvent, 0)
	for rows.Next() {
		var e model.OutboxEvent
		if err = rows.Scan(
			&e.Id, &e.Subject, &e.Payload,
			&e.Status, &e.CreatedAt, &e.AvailableAt,
			&e.LockedBy, &e.LockedUntil,
			&e.Attempts, &e.LastError,
		); err != nil {
			return nil, fmt.Errorf("failed to scan outbox_event: %w", err)
		}
		events = append(events, e)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}
	return events, nil
}

func (store *OutboxStore) MarkCompleted(ctx context.Context, tx *sql.Tx, workerId, id string) error {
	if id == "" {
		return fmt.Errorf("outbox event id is required")
	}
	if workerId == "" {
		return fmt.Errorf("workerId is required")
	}
	res, err := store.execContext(ctx, tx, `
	UPDATE outbox_events
	SET status = ?,
		locked_by = NULL,
		locked_until = NULL,
		last_error = NULL
	WHERE id = ? AND locked_by = ?;
	`, model.OutboxStatusCompleted, id, workerId)
	if err != nil {
		return fmt.Errorf("failed to mark outbox event completed: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("no outbox event updated")
	}
	return nil
}

func (store *OutboxStore) Requeue(ctx context.Context, tx *sql.Tx, workerId, id string, availableAt int64, lastError string) error {
	if id == "" {
		return fmt.Errorf("outbox event id is required")
	}
	if workerId == "" {
		return fmt.Errorf("workerId is required")
	}
	if availableAt == 0 {
		availableAt = time.Now().Unix()
	}
	res, err := store.execContext(ctx, tx, `
	UPDATE outbox_events
	SET status = ?,
		available_at = ?,
		locked_by = NULL,
		locked_until = NULL,
		last_error = ?
	WHERE id = ? AND locked_by = ?;
	`, model.OutboxStatusNew, availableAt, lastError, id, workerId)
	if err != nil {
		return fmt.Errorf("failed to requeue outbox event: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("no outbox event updated")
	}
	return nil
}

func (store *OutboxStore) MarkFailed(ctx context.Context, tx *sql.Tx, workerId, id string, lastError string) error {
	if id == "" {
		return fmt.Errorf("outbox event id is required")
	}
	if workerId == "" {
		return fmt.Errorf("workerId is required")
	}
	res, err := store.execContext(ctx, tx, `
	UPDATE outbox_events
	SET status = ?,
		locked_by = NULL,
		locked_until = NULL,
		last_error = ?
	WHERE id = ? AND locked_by = ?;
	`, model.OutboxStatusFailed, lastError, id, workerId)
	if err != nil {
		return fmt.Errorf("failed to mark outbox event failed: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("no outbox event updated")
	}
	return nil
}

func initOutboxEventTable(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS outbox_events (
		id TEXT PRIMARY KEY,                -- UUID string
		subject TEXT NOT NULL,
		payload BLOB NOT NULL,              -- JSON blob

		status INTEGER NOT NULL DEFAULT 0,  -- 0=NEW, 1=PROCESSING, 2=COMPLETED, 3=FAILED

		created_at INTEGER NOT NULL,        -- unix seconds
		available_at INTEGER NOT NULL,      -- unix seconds

		locked_by TEXT NULL,
		locked_until INTEGER NULL,          -- unix seconds

		attempts INTEGER NOT NULL DEFAULT 0,
		last_error TEXT NULL
	);`)
	if err != nil {
		return fmt.Errorf("failed to create outbox_events: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `
	CREATE INDEX IF NOT EXISTS idx_outbox_ready
	ON outbox_events (status, available_at, created_at);
	`); err != nil {
		return fmt.Errorf("failed to create idx_outbox_ready: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
	CREATE INDEX IF NOT EXISTS idx_outbox_processing_lease
	ON outbox_events (status, locked_until);
	`); err != nil {
		return fmt.Errorf("failed to create idx_outbox_processing_lease: %w", err)
	}
	return nil
}

func (store *OutboxStore) execContext(ctx context.Context, tx *sql.Tx, query string, args ...any) (sql.Result, error) {
	if tx != nil {
		return tx.ExecContext(ctx, query, args...)
	}
	return store.db.ExecContext(ctx, query, args...)
}

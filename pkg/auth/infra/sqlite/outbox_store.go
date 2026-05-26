package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/elug3/gochat/pkg/auth/domain"
)

var errOutboxEventNotFound = errors.New("outbox event not found")

// OutboxStore provides sqlite persistence operations for auth outbox events.
type OutboxStore struct {
	db *sql.DB
}

func NewOutboxStore(db *sql.DB) (*OutboxStore, error) {
	if db == nil {
		return nil, fmt.Errorf("db is required")
	}
	return &OutboxStore{db: db}, nil
}

func (s *OutboxStore) Insert(ctx context.Context, row domain.OutboxRecord) error {
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

	_, err := s.db.ExecContext(ctx, `
	INSERT INTO outbox_events (
		id, subject, payload, created_at, available_at, attempts
	) VALUES (?, ?, ?, ?, ?, ?);
	`, row.Id, row.Subject, row.Payload, createdAt, availableAt, attempts)
	if err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}
	return nil
}

func (s *OutboxStore) ClaimBatch(ctx context.Context, workerID string, limit, leaseSeconds int) ([]domain.OutboxEvent, error) {
	if workerID == "" {
		return nil, fmt.Errorf("workerID is required")
	}
	if limit <= 0 {
		return []domain.OutboxEvent{}, nil
	}
	if leaseSeconds <= 0 {
		leaseSeconds = 1
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
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
			(status = ? AND available_at <= %s)
			OR
			(status = ? AND locked_until IS NOT NULL AND locked_until <= %s)
		ORDER BY created_at
		LIMIT ?
	);
	`, nowExpr, nowExpr, nowExpr)

	_, err = tx.ExecContext(
		ctx,
		claimSQL,
		domain.OutboxStatusProcessing,
		workerID,
		leaseSeconds,
		domain.OutboxStatusNew,
		domain.OutboxStatusProcessing,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("claim outbox events: %w", err)
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

	rows, err := tx.QueryContext(ctx, selectSQL, domain.OutboxStatusProcessing, workerID, limit)
	if err != nil {
		return nil, fmt.Errorf("query claimed events: %w", err)
	}
	defer rows.Close()

	events := make([]domain.OutboxEvent, 0, limit)
	for rows.Next() {
		var event domain.OutboxEvent
		if err = rows.Scan(
			&event.Id,
			&event.Subject,
			&event.Payload,
			&event.Status,
			&event.CreatedAt,
			&event.AvailableAt,
			&event.LockedBy,
			&event.LockedUntil,
			&event.Attempts,
			&event.LastError,
		); err != nil {
			return nil, fmt.Errorf("scan outbox event: %w", err)
		}
		events = append(events, event)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate claimed events: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit claim batch: %w", err)
	}
	return events, nil
}

func (s *OutboxStore) MarkCompleted(ctx context.Context, tx *sql.Tx, workerID, id string) error {
	if id == "" {
		return fmt.Errorf("outbox event id is required")
	}
	if workerID == "" {
		return fmt.Errorf("workerID is required")
	}

	exec := s.db.ExecContext
	if tx != nil {
		exec = tx.ExecContext
	}

	res, err := exec(ctx, `
	UPDATE outbox_events
	SET status = ?,
		locked_by = NULL,
		locked_until = NULL,
		last_error = NULL
	WHERE id = ? AND locked_by = ?;
	`, domain.OutboxStatusCompleted, id, workerID)
	if err != nil {
		return fmt.Errorf("mark completed: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("mark completed rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errOutboxEventNotFound
	}
	return nil
}

func (s *OutboxStore) Requeue(ctx context.Context, tx *sql.Tx, workerID, id string, availableAt int64, lastError string) error {
	if id == "" {
		return fmt.Errorf("outbox event id is required")
	}
	if workerID == "" {
		return fmt.Errorf("workerID is required")
	}
	if availableAt == 0 {
		availableAt = time.Now().Unix()
	}

	exec := s.db.ExecContext
	if tx != nil {
		exec = tx.ExecContext
	}

	res, err := exec(ctx, `
	UPDATE outbox_events
	SET status = ?,
		available_at = ?,
		locked_by = NULL,
		locked_until = NULL,
		last_error = ?
	WHERE id = ? AND locked_by = ?;
	`, domain.OutboxStatusNew, availableAt, lastError, id, workerID)
	if err != nil {
		return fmt.Errorf("requeue event: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("requeue rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errOutboxEventNotFound
	}
	return nil
}

func (s *OutboxStore) MarkFailed(ctx context.Context, tx *sql.Tx, workerID, id, lastError string) error {
	if id == "" {
		return fmt.Errorf("outbox event id is required")
	}
	if workerID == "" {
		return fmt.Errorf("workerID is required")
	}

	exec := s.db.ExecContext
	if tx != nil {
		exec = tx.ExecContext
	}

	res, err := exec(ctx, `
	UPDATE outbox_events
	SET status = ?,
		locked_by = NULL,
		locked_until = NULL,
		last_error = ?
	WHERE id = ? AND locked_by = ?;
	`, domain.OutboxStatusFailed, lastError, id, workerID)
	if err != nil {
		return fmt.Errorf("mark failed: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("mark failed rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errOutboxEventNotFound
	}
	return nil
}

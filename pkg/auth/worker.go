package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/elug3/gochat/pkg/auth/internal/model"
	"github.com/elug3/gochat/pkg/auth/internal/store/sqlite3"
	"github.com/elug3/gochat/shared/events"
	"github.com/google/uuid"
)

const (
	defaultBatchSize    = 10
	defaultLease        = 30 * time.Second
	defaultPollInterval = 500 * time.Millisecond
	defaultMaxAttempts  = 5
	defaultRetryDelay   = 5 * time.Second
)

type WorkerConfig struct {
	Outbox       *sqlite3.OutboxStore
	Publisher    *events.Publisher
	ID           string
	BatchSize    int
	Lease        time.Duration
	PollInterval time.Duration
	MaxAttempts  int
	RetryDelay   time.Duration
}

type worker struct {
	outbox       *sqlite3.OutboxStore
	publisher    *events.Publisher
	id           string
	batchSize    int
	lease        time.Duration
	pollInterval time.Duration
	maxAttempts  int
	retryDelay   time.Duration
}

func NewWorker(cfg WorkerConfig) (*worker, error) {
	if cfg.Outbox == nil {
		return nil, fmt.Errorf("%w: outbox store is required", ErrInvalidConfig)
	}
	if cfg.Publisher == nil {
		return nil, fmt.Errorf("%w: publisher is required", ErrInvalidConfig)
	}
	id := cfg.ID
	if id == "" {
		id = uuid.NewString()
	}
	batchSize := cfg.BatchSize
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}
	lease := cfg.Lease
	if lease <= 0 {
		lease = defaultLease
	}
	pollInterval := cfg.PollInterval
	if pollInterval <= 0 {
		pollInterval = defaultPollInterval
	}
	maxAttempts := cfg.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = defaultMaxAttempts
	}
	retryDelay := cfg.RetryDelay
	if retryDelay <= 0 {
		retryDelay = defaultRetryDelay
	}
	return &worker{
		outbox:       cfg.Outbox,
		publisher:    cfg.Publisher,
		id:           id,
		batchSize:    batchSize,
		lease:        lease,
		pollInterval: pollInterval,
		maxAttempts:  maxAttempts,
		retryDelay:   retryDelay,
	}, nil
}

func (w *worker) Publish(ctx context.Context, row *model.OutboxRecord) error {
	if w == nil {
		return fmt.Errorf("worker is nil")
	}
	if row == nil {
		return fmt.Errorf("outbox record is required")
	}
	return w.outbox.Insert(ctx, nil, row)
}

func (w *worker) Run(ctx context.Context) {
	if w == nil {
		return
	}
	leaseSeconds := int(w.lease.Seconds())
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		rows, err := w.outbox.ClaimBatch(ctx, w.id, w.batchSize, leaseSeconds)
		if err != nil {
			w.sleep(ctx)
			continue
		}
		if len(rows) == 0 {
			w.sleep(ctx)
			continue
		}
		for _, row := range rows {
			if ctx.Err() != nil {
				return
			}
			if err := w.publisher.Publish(rawEvent{
				subject: row.Subject,
				payload: row.Payload,
			}); err != nil {
				w.handleFailure(ctx, row, err)
				continue
			}
			if err := w.outbox.MarkCompleted(ctx, nil, w.id, row.Id); err != nil {
				w.sleep(ctx)
			}
		}
	}
}

func (w *worker) handleFailure(ctx context.Context, row model.OutboxEvent, publishErr error) {
	if w == nil || publishErr == nil {
		return
	}
	errMsg := publishErr.Error()
	if row.Attempts >= w.maxAttempts {
		_ = w.outbox.MarkFailed(ctx, nil, w.id, row.Id, errMsg)
		return
	}
	nextAttempt := time.Now().Add(w.retryDelay).Unix()
	_ = w.outbox.Requeue(ctx, nil, w.id, row.Id, nextAttempt, errMsg)
}

func (w *worker) sleep(ctx context.Context) {
	if w.pollInterval <= 0 {
		return
	}
	timer := time.NewTimer(w.pollInterval)
	defer timer.Stop()

	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

type rawEvent struct {
	subject string
	payload []byte
}

func (e rawEvent) Subject() string {
	return e.subject
}

func (e rawEvent) MarshalJSON() ([]byte, error) {
	if len(e.payload) == 0 {
		return []byte("null"), nil
	}
	return e.payload, nil
}

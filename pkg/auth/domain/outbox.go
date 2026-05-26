package domain

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/elug3/gochat/shared/events"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

// OutboxStatus represents the processing status of an outbox event
type OutboxStatus int

const (
	OutboxStatusNew OutboxStatus = iota
	OutboxStatusProcessing
	OutboxStatusCompleted
	OutboxStatusFailed
)

// OutboxRecord represents a record to be stored in the outbox table
type OutboxRecord struct {
	Id          string
	Subject     string
	Payload     []byte
	CreatedAt   int64
	AvailableAt int64
	Attempts    int
}

// OutboxEvent represents an event retrieved from the outbox with full status details
type OutboxEvent struct {
	Id          string
	Subject     string
	Payload     []byte
	Status      OutboxStatus
	CreatedAt   int64
	AvailableAt int64
	LockedBy    string
	LockedUntil int64
	Attempts    int
	LastError   string
}

// NewOutboxRecord creates a new OutboxRecord from an event for storage
func NewOutboxRecord(e events.Event) (*OutboxRecord, error) {
	if e == nil {
		return nil, fmt.Errorf("event cannot be nil")
	}

	payload, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("marshal event: %w", err)
	}

	now := time.Now().Unix()
	return &OutboxRecord{
		Id:          uuid.NewString(),
		Subject:     e.Subject(),
		Payload:     payload,
		CreatedAt:   now,
		AvailableAt: now,
		Attempts:    0,
	}, nil
}

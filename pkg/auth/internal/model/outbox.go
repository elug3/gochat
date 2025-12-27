package model

import (
	"encoding/json"
	"time"

	"github.com/elug3/gochat/shared/events"
	"github.com/google/uuid"
)

type OutboxRecord struct {
	Id          string
	Subject     string
	Payload     []byte
	CreatedAt   int64
	AvailableAt int64
	Attempts    int
}

type OutboxStatus int

const (
	OutboxStatusNew OutboxStatus = iota
	OutboxStatusProcessing
	OutboxStatusCompleted
	OutboxStatusFailed
)

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

func NewOutboxRecord(e events.Event) (*OutboxRecord, error) {
	payload, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return &OutboxRecord{
		Id:          uuid.NewString(),
		Subject:     e.Subject(),
		Payload:     payload,
		CreatedAt:   time.Now().Unix(),
		AvailableAt: time.Now().Unix(),
		Attempts:    0,
	}, nil
}

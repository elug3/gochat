package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
)

type Subscriber struct {
	DLQ *DLQ
	sub *nats.Subscription

	handlers map[string]Handler
	logger   *zerolog.Logger
	mu       sync.RWMutex
}

type Handler interface {
	Handle(ctx context.Context, subject string, data []byte) error
}

type DLQ struct {
	js               nats.JetStreamContext
	originalStream   string
	originalConsumer string
}
type DLQEntry struct {
	OriginalSubject string
	OriginalStream  string
	Consumer        string
	Err             string
	Timestamp       time.Time
	Payload         []byte
}

var (
	ErrNoHandlerFound = errors.New("no handler found for event")
	ErrHandlerFailed  = errors.New("handler failed to process event")
)

type disposition string

const (
	dispAck  = "ack"
	dispNak  = "nak"
	dispTerm = "term"
	dispDLQ  = "dlq"
)

var (
	// timeouts for subscriber operations
	FetchTimeout = 5 * time.Second
	// timeouts for message acknowledge operations
	AcknowledgeTimeout = 10 * time.Second
	// timeouts for handler execution
	HandlerTimeout = 10 * time.Second
)

type HandlerFunc func(ctx context.Context, subject string, data []byte) error

func (fn HandlerFunc) Handle(ctx context.Context, subject string, data []byte) error {
	return fn(ctx, subject, data)
}

func NewSubscriber(nc *nats.Conn, logger *zerolog.Logger, stream string, cfg *nats.ConsumerConfig) (*Subscriber, error) {
	if nc == nil {
		return nil, fmt.Errorf("nats connection cannot be nil")
	}
	if stream == "" {
		return nil, fmt.Errorf("stream cannot be empty")
	}
	if cfg == nil {
		return nil, fmt.Errorf("consumer config cannot be nil")
	}
	if cfg.Durable == "" {
		return nil, fmt.Errorf("consumer durable name cannot be empty")
	}
	if logger == nil {
		nop := zerolog.Nop()
		logger = &nop
	}
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to get JetStream context: %w", err)
	}

	// if not exists, create consumer
	if _, err = js.ConsumerInfo(stream, cfg.Durable); err != nil {
		if !errors.Is(err, nats.ErrConsumerNotFound) {
			return nil, fmt.Errorf("failed to get stream info: %w", err)
		}

		// create consumer
		if _, err = js.AddConsumer(stream, cfg); err != nil {
			if _, infoErr := js.ConsumerInfo(stream, cfg.Durable); infoErr != nil {
				if errors.Is(infoErr, nats.ErrConsumerNotFound) {
					return nil, fmt.Errorf("failed to create consumer %s: %w", cfg.Durable, err)
				}
				return nil, fmt.Errorf("failed to verify consumer %s after create error: %w", cfg.Durable, infoErr)
			}
			logger.Warn().
				Err(err).
				Msgf("consumer %s already exists; continuing", cfg.Durable)
		}
	}
	sub, err := js.PullSubscribe(
		"", // subject ignored when binding to a durble consumer
		cfg.Durable,
		nats.Bind(stream, cfg.Durable),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull subscription: %w", err)
	}

	DLQ, err := NewDLQ(js, stream, cfg.Durable)
	if err != nil {
		return nil, fmt.Errorf("failed to create DLQ: %w", err)
	}

	return &Subscriber{
		DLQ:      DLQ,
		handlers: make(map[string]Handler),
		sub:      sub,
		logger:   logger,
	}, nil
}

func (s *Subscriber) Handle(pattern string, handler Handler) {
	if handler == nil {
		panic("handler cannot be nil")
	}
	if err := validatePattern(pattern); err != nil {
		panic(fmt.Sprintf("invalid pattern %q: %v", pattern, err))
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.handlers[pattern]; exists {
		panic(fmt.Sprintf("handler already exists for pattern %q", pattern))
	}
	s.handlers[pattern] = handler
}

func validatePattern(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("pattern cannot be empty")
	}
	parts := strings.Split(pattern, ".")
	for i, part := range parts {
		if part == "" {
			return fmt.Errorf("pattern cannot have empty parts")
		}
		if part == ">" && i != len(parts)-1 {
			return fmt.Errorf("pattern '>' must be the last part")
		}
	}
	return nil
}

// HandleFunc registers a handler function for the given subject pattern.
func (s *Subscriber) HandleFunc(pattern string, handlerFunc func(ctx context.Context, subject string, data []byte) error) {
	s.Handle(pattern, HandlerFunc(handlerFunc))
}

func (s *Subscriber) getHandler(subject string) (Handler, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if h, exists := s.handlers[subject]; exists {
		return h, true
	}

	minPrio := -1
	var selected Handler
	for pattern, h := range s.handlers {
		if match, prio := wildcardMatch(subject, pattern); match {
			if minPrio == -1 && prio >= 0 {
				// matched first handler
				minPrio = prio
				selected = h
			} else if prio < minPrio {
				minPrio = prio
				selected = h
			}
		}
	}
	return selected, selected != nil
}

func (s *Subscriber) Run(ctx context.Context) error {
	defer func() {
		if err := s.sub.Drain(); err != nil && !errors.Is(err, nats.ErrConnectionClosed) && !errors.Is(err, nats.ErrBadSubscription) {
			s.logger.Warn().Err(err).Msg("failed to drain subscription")
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		fetchCtx, fetchCancel := context.WithTimeout(ctx, FetchTimeout)
		msgs, err := s.sub.Fetch(10, nats.Context(fetchCtx))
		fetchCancel()

		if err != nil {
			switch {
			case errors.Is(err, context.DeadlineExceeded),
				errors.Is(err, nats.ErrTimeout):
				continue
			case errors.Is(err, context.Canceled):
				return nil
			default:
				s.logger.Warn().Err(err).Msg("failed to fetch messages")
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(3 * time.Second):
				}
				continue
			}
		}

		for _, msg := range msgs {
			select {
			case <-ctx.Done():
				return nil
			default:
			}
			msgCtx, msgCancel := context.WithTimeout(ctx, HandlerTimeout)
			disp, hErr := s.handleMessage(msgCtx, msg)
			if hErr != nil {
				s.logger.Warn().
					Str("subject", msg.Subject).
					Err(hErr).
					Msg("failed to handle message")
			}
			msgCancel()

			ackCtx, ackCancel := context.WithTimeout(ctx, AcknowledgeTimeout)
			var err error
			switch disp {
			case dispAck:
				err = msg.Ack(nats.Context(ackCtx))
			case dispNak:
				err = msg.NakWithDelay(3*time.Second, nats.Context(ackCtx))
			case dispTerm:
				err = msg.Term(nats.Context(ackCtx))
			case dispDLQ:
				if err = s.DLQ.Publish(ackCtx, msg, hErr); err != nil {
					s.logger.Warn().
						Str("subject", msg.Subject).
						Err(err).
						Msg("failed to publish message to DLQ")
				} else {
					err = msg.Term(nats.Context(ackCtx))
				}
			default: // unknown disposition, treat as DLQ
				if err = s.DLQ.Publish(ackCtx, msg, nil); err != nil {
					s.logger.Warn().
						Str("subject", msg.Subject).
						Err(err).
						Msg("failed to publish message to DLQ")
				} else {
					err = msg.Term(nats.Context(ackCtx))
				}
			}
			ackCancel()

			if err == nil {
				s.logger.Info().
					Str("disposition", string(disp)).
					Str("subject", msg.Subject).
					Msg("acknowledged message")
			} else {
				// if failed to acknowledge message, log it
				s.logger.Warn().
					Str("disposition", string(disp)).
					Str("subject", msg.Subject).
					Err(err).
					Msg("failed to acknowledge message")
			}
		}
	}
}

func (s *Subscriber) Close() error {
	if s.sub == nil {
		return nil
	}
	return s.sub.Drain()
}

// handleMessage processes a single nats message and returns the disposition and any error encountered.
// if no handler is found, returns dispDLQ and ErrNoHandlerFound.
// if handler fails, returns dispNak and ErrHandlerFailed.
func (s *Subscriber) handleMessage(ctx context.Context, msg *nats.Msg) (disposition, error) {
	h, exists := s.getHandler(msg.Subject)
	if !exists {
		return dispDLQ, ErrNoHandlerFound
	}

	err := h.Handle(ctx, msg.Subject, msg.Data)
	if err != nil {
		return dispNak, fmt.Errorf("%w: %v", ErrHandlerFailed, err)
	}

	return dispAck, nil
}

// NewDLQ creates a new DLQ instance.
// only one DLQ per stream and consumer should be created.
func NewDLQ(js nats.JetStreamContext, originalStream, originalConsumer string) (*DLQ, error) {
	if _, err := js.AddStream(&nats.StreamConfig{
		Name:     "DLQ",
		Subjects: []string{"DLQ.>"},
	}); err != nil {
		if _, infoErr := js.StreamInfo("DLQ"); infoErr != nil {
			if errors.Is(infoErr, nats.ErrStreamNotFound) {
				// failed to create stream and it still doesn't exist
				return nil, fmt.Errorf("failed to create DLQ stream: %w", err)
			}
			return nil, fmt.Errorf("failed to verify DLQ stream: %w", infoErr)
		}
	}
	return &DLQ{
		js:               js,
		originalStream:   originalStream,
		originalConsumer: originalConsumer,
	}, nil

}

// Publish publishes a message to the DLQ with the given error.
func (dlq *DLQ) Publish(ctx context.Context, msg *nats.Msg, err error) error {
	subj := fmt.Sprintf("DLQ.%s", msg.Subject)
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}

	entry := DLQEntry{
		OriginalSubject: msg.Subject,
		OriginalStream:  dlq.originalStream,
		Consumer:        dlq.originalConsumer,
		Err:             errMsg,
		Timestamp:       time.Now().UTC(),
		Payload:         msg.Data,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal DLQ entry: %w", err)
	}
	_, err = dlq.js.PublishMsg(&nats.Msg{
		Subject: subj,
		Data:    data,
	}, nats.Context(ctx))
	if err != nil {
		return fmt.Errorf("failed to publish message to DLQ: %w", err)
	}
	return nil
}

// wildcardMatch checks if the subject matches the pattern with wildcards.
// returns match (bool) and priority (int).
// lower priority value means more specific match.
// if all parts exact matchs, prio is zero.
// earlier wildcards have higher weight than later wildcards.
func wildcardMatch(subject, pattern string) (match bool, prio int) {
	subParts := strings.Split(subject, ".")
	patParts := strings.Split(pattern, ".")

	base := 3 // literal, '*', '>'
	for _, part := range subParts {
		if part == "" {
			return false, -1
		}
	}

	for i, part := range patParts {
		if part == "" {
			return false, -1
		}
		if part == ">" {
			if i != len(patParts)-1 {
				return false, -1
			}
			if i >= len(subParts) {
				return false, -1
			}
			for j := i; j < len(subParts); j++ {
				prio += tokenScore(base, 2, len(subParts)-1-j)
			}
			return true, prio
		}
		if i >= len(subParts) {
			return false, -1
		}

		switch part {
		case "*":
			prio += tokenScore(base, 1, len(subParts)-1-i)
		default:
			if part != subParts[i] {
				return false, -1
			}
		}
	}

	if len(patParts) != len(subParts) {
		return false, -1
	}
	return true, prio
}

func tokenScore(base, digit, i int) int {
	// base = 3 (literal, '*', '>')
	// n = total parts ("a.b.c" = 3)
	// i = remaining index (0-based from the right)
	// digit = token specificity value ( literal=0, '*'=1, '>'=2 )
	// weight = base^i
	// score = digit * weight
	pow := func(b int, exp int) int {
		result := 1
		for exp > 0 {
			result *= b
			exp--
		}
		return result
	}

	weight := pow(base, i)

	return digit * weight
}

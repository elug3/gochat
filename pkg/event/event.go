package event

import (
	"context"
	"fmt"
	"sync"
)

type EventHandler struct {
	subjects map[string]*Subject
	mu       sync.RWMutex
}

type Event struct {
	Data interface{}
}

type Subject struct {
	channels []chan interface{}
	once     sync.Once
}

type Subscriber struct {
	ch     chan interface{}
	cancel context.CancelFunc
}

func NewSubscriber(ch chan interface{}) *Subscriber {
	subc := &Subscriber{ch: ch}
	return subc
}

func NewEventHandler() *EventHandler {
	eh := &EventHandler{
		subjects: make(map[string]*Subject),
	}
	return eh
}

func NewSubject() *Subject {
	subj := &Subject{
		channels: make([]chan interface{}, 0),
	}
	return subj
}

func (eh *EventHandler) Register(ctx context.Context, pattern string, fn func(e *Event) error) error {
	subj := eh.getSubject(pattern, true)
	subc := subj.createSubscriber()

	go subc.Subscribe(ctx, fn)
	return nil
}

func (eh *EventHandler) Publish(pattern string, data interface{}) error {
	subj := eh.getSubject(pattern, true)
	if subj == nil {
		return fmt.Errorf("no subject")
	}
	subj.Publish(data)
	return nil

}

// getSubject returnss the subject for the given pattern.
// If it does not exist, a new one is createed and returned
func (eh *EventHandler) getSubject(pattern string, create bool) *Subject {
	subj, exist := eh.subjects[pattern]
	if !exist && create {
		subj = NewSubject()
		eh.subjects[pattern] = subj
	}
	return subj
}

func (eh *EventHandler) deleteSubject(pattern string) {
}

func (subj *Subject) Publish(data interface{}) {
	for _, ch := range subj.channels {
		ch <- data
	}
}

func (subj *Subject) createSubscriber() *Subscriber {
	ch := make(chan interface{}, 10)
	subj.channels = append(subj.channels, ch)
	subc := NewSubscriber(ch)

	return subc
}

// TODO
func (subj *Subject) deleteSubscriber() {
}

// TODO: add Done() receiver
func (subc *Subscriber) Subscribe(ctx context.Context, fn func(e *Event) error) error {
	ctx, cancel := context.WithCancel(ctx)
	subc.cancel = cancel

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case data, ok := <-subc.ch:
			fmt.Println("ok:", ok)
			if err := fn(&Event{Data: data}); err != nil {
				return err
			}
		}
	}
}

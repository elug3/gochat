package events

import (
	"encoding/json"
	"fmt"
)

type Event interface {
	Subject() string
}

const (
	// SubjectUserRegistered is the subject for UserRegistered events.
	SubjectUserRegistered = "auth.registered"

	SubjectGroupCreated = "contacts.group.created"
	SubjectGroupDeleted = "contacts.group.deleted"
	SubjectMemberJoined = "contacts.member.joined"
	SubjectMemberLeft   = "contacts.member.left"
	SubjectMessageSent  = "message.sent"
	SubjectMessageRead  = "message.read"

	SubjectWebsocketSent = "websocket.sent"
)

const (
	SubjectUserAll      = "user.>"
	SubjectContactsAll  = "contacts.>"
	SubjectMessageAll   = "message.>"
	SubjectWebsocketAll = "websocket.>"
	SubjectAuthAll      = "auth.>"
)

const (
	StreamAuth      = "AUTH"
	StreamContacts  = "CONTACTS"
	StreamMessages  = "MESSAGES"
	StreamWebsocket = "WEBSOCKET"
)

// UserRegistered is published by AuthService after credentials are created for a user.
type UserRegistered struct {
	UserId    int32
	Username  string
	Timestamp int64
}

type GroupCreated struct {
	GroupId   int
	GroupName string
	TimeStamp int64
}

type GroupDeleted struct {
	GroupId   int
	TimeStamp int64
}

type MemberJoined struct {
	GroupId   int
	UserId    int32
	TimeStamp int64
}

type MemberLeft struct {
	GroupId   int
	UserId    int32
	TimeStamp int64
}

type MessageSent struct {
	ChatId    int
	SenderId  int32
	Content   string
	TimeStamp int64
}

type MessageRead struct {
	ChatId    int
	UserId    int32
	TimeStamp int64
}

type WebsocketSent struct {
	ChatId    int
	SenderId  int32
	Content   string
	TimeStamp int64
}

func (e UserRegistered) Subject() string {
	return SubjectUserRegistered
}

func (e GroupCreated) Subject() string {
	return SubjectGroupCreated
}

func (e GroupDeleted) Subject() string {
	return SubjectGroupDeleted
}

func (e MemberJoined) Subject() string {
	return SubjectMemberJoined
}

func (e MemberLeft) Subject() string {
	return SubjectMemberLeft
}

func (e MessageSent) Subject() string {
	return SubjectMessageSent
}

func (e WebsocketSent) Subject() string {
	return SubjectWebsocketSent
}

func (e MessageRead) Subject() string {
	return SubjectMessageRead
}

type EventFactory func() Event

var subjects = map[string]EventFactory{
	SubjectUserRegistered: func() Event { return &UserRegistered{} },
	SubjectGroupCreated:   func() Event { return &GroupCreated{} },
	SubjectGroupDeleted:   func() Event { return &GroupDeleted{} },
	SubjectMemberJoined:   func() Event { return &MemberJoined{} },
	SubjectMemberLeft:     func() Event { return &MemberLeft{} },
	SubjectMessageSent:    func() Event { return &MessageSent{} },
	SubjectWebsocketSent:  func() Event { return &WebsocketSent{} },
	SubjectMessageRead:    func() Event { return &MessageRead{} },
}

func UnmarshalEvent(subject string, data []byte) (Event, error) {
	factory, ok := subjects[subject]
	if !ok {
		return nil, fmt.Errorf("unknown subject: %s", subject)
	}
	event := factory()
	if err := json.Unmarshal(data, event); err != nil {
		return nil, err
	}
	return event, nil
}

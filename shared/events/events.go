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

	SubjectGroupCreated  = "contacts.group.created"
	SubjectGroupDeleted  = "contacts.group.deleted"
	SubjectMemberJoined  = "contacts.member.joined"
	SubjectMemberLeft    = "contacts.member.left"
	SubjectContactsReset = "contacts.reset"
	SubjectMessageSent   = "message.sent"
	SubjectMessageRead   = "message.read"

	SubjectWebsocketSent         = "websocket.sent"
	SubjectWebsocketConnected    = "websocket.connected"
	SubjectWebsocketDisconnected = "websocket.disconnected"
)

const (
	SubjectUserAll      = "user.>"
	SubjectContactsAll  = "contacts.>"
	SubjectMessageAll   = "message.>"
	SubjectWebsocketAll = "websocket.>"
	SubjectAuthAll      = "auth.>"
)

var (
	// message service consumes websocket events to send messages
	SubjectsWebsocketMessage = []string{SubjectWebsocketSent}
)

const (
	StreamAuth      = "AUTH"
	StreamContacts  = "CONTACTS"
	StreamMessages  = "MESSAGES"
	StreamWebsocket = "WEBSOCKET"
	StreamPresence  = "PRESENCE"
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

type ContactsReset struct {
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

type WebsocketConnected struct {
	UserId    int32
	TimeStamp int64
	IsFirst   bool // true if this is the first connection for the user
}

type WebsocketDisconnected struct {
	UserId    int32
	TimeStamp int64
	IsLast    bool // true if this is the last connection for the user
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

func (ContactsReset) Subject() string {
	return SubjectContactsReset
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

func (WebsocketConnected) Subject() string {
	return SubjectWebsocketConnected
}

func (WebsocketDisconnected) Subject() string {
	return SubjectWebsocketDisconnected
}

type EventFactory func() Event

var subjects = map[string]EventFactory{
	SubjectUserRegistered:        func() Event { return &UserRegistered{} },
	SubjectGroupCreated:          func() Event { return &GroupCreated{} },
	SubjectGroupDeleted:          func() Event { return &GroupDeleted{} },
	SubjectMemberJoined:          func() Event { return &MemberJoined{} },
	SubjectMemberLeft:            func() Event { return &MemberLeft{} },
	SubjectContactsReset:         func() Event { return &ContactsReset{} },
	SubjectMessageSent:           func() Event { return &MessageSent{} },
	SubjectWebsocketSent:         func() Event { return &WebsocketSent{} },
	SubjectMessageRead:           func() Event { return &MessageRead{} },
	SubjectWebsocketConnected:    func() Event { return &WebsocketConnected{} },
	SubjectWebsocketDisconnected: func() Event { return &WebsocketDisconnected{} },
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

package events

import "net"

type Event interface {
	Subject() string
}

const (
	// SubjectUserRegistered is the subject for UserRegistered events.
	SubjectUserRegistered = "auth.registered"
	SubjectUserLoggedIn   = "auth.logged_in"

	SubjectProfileCreated = "contacts.profile.created"
	SubjectGroupCreated   = "contacts.group.created"
	SubjectGroupDeleted   = "contacts.group.deleted"
	SubjectMemberJoined   = "contacts.member.joined"
	SubjectMemberLeft     = "contacts.member.left"
	SubjectContactsReset  = "contacts.reset"
	SubjectMessageSent    = "message.sent"
	SubjectMessageRead    = "message.read"

	SubjectWebsocketSent         = "websocket.sent"
	SubjectWebsocketConnected    = "websocket.connected"
	SubjectWebsocketDisconnected = "websocket.disconnected"
)

var (
	SubjectContactsGroupAll  = "contacts.group.*"
	SubjectContactsMemberAll = "contacts.member.*"
	SubjectMessageAll        = "message.*"
	SubjectWebsocketAll      = "websocket.*"
)

var (
	APP_STREAM = "APP_STREAM"
)

var (
	DurableChatView  = "CHATVIEW"
	DurableContacts  = "CONTACTS"
	DurableMessage   = "MESSAGE"
	DurableWebsocket = "WEBSOCKET"
	DurablePresence  = "PRESENCE"
	DurableAvatar    = "AVATAR"
)

var (
	// message service consumes websocket events to send messages
	SubjectsWebsocketMessage = []string{SubjectWebsocketSent}
)

// UserRegistered is published by AuthService after credentials are created for a user.
type UserRegistered struct {
	UserId    int32
	Username  string
	Name      string
	Timestamp int64
}

// UserLoggedIn is published by AuthService when a user successfully logs in.
type UserLoggedIn struct {
	UserId    int32
	Username  string
	IP        net.IP
	UserAgent string
	Timestamp int64
}

type ProfileCreated struct {
	UserId    int32
	Timestamp int64
}

type GroupCreated struct {
	GroupId   int
	GroupName string
	Timestamp int64
}

type GroupDeleted struct {
	GroupId   int
	Timestamp int64
}

type MemberJoined struct {
	GroupId   int
	UserId    int32
	Timestamp int64
}

type MemberLeft struct {
	GroupId   int
	UserId    int32
	Timestamp int64
}

type ContactsReset struct {
	Timestamp int64
}

type MessageSent struct {
	ChatId    int
	SenderId  int32
	Content   string
	Timestamp int64
}

type MessageRead struct {
	ChatId    int
	UserId    int32
	Timestamp int64
}

type WebsocketSent struct {
	ChatId    int
	SenderId  int32
	Content   string
	Timestamp int64
}

type WebsocketConnected struct {
	UserId    int32
	Timestamp int64
	IsFirst   bool // true if this is the first connection for the user
}

type WebsocketDisconnected struct {
	UserId    int32
	Timestamp int64
	IsLast    bool // true if this is the last connection for the user
}

func (e UserRegistered) Subject() string {
	return SubjectUserRegistered
}

func (e UserLoggedIn) Subject() string {
	return SubjectUserLoggedIn
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

func (ProfileCreated) Subject() string {
	return SubjectProfileCreated
}

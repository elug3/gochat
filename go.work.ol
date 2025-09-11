go 1.24.6

use (
	.
	./cmd/gochat-auth_server
	./cmd/gochat-chat_projection_server
	./cmd/gochat-chatview-server
	./cmd/gochat-contacts_server
	./cmd/gochat-message_server
	./pkg/auth
	./pkg/chatview/chat-projection
	./pkg/chatview/chatquery
	./pkg/contacts
	./pkg/message
	./pkg/users
)

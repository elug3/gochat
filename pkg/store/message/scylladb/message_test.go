package scylladb

import (
	"testing"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

func NewTestStore(t *testing.T) (*MessageStore, error) {
	store, err := NewMessageStore("messages_test", "node-0.gce-us-east-1.bf75b9cac053869fd78b.clusters.scylla.cloud", "node-1.gce-us-east-1.bf75b9cac053869fd78b.clusters.scylla.cloud", "node-2.gce-us-east-1.bf75b9cac053869fd78b.clusters.scylla.cloud")
	if err != nil {
		return nil, err
	}
	return store, nil
}

// func TestMessage_Create(t *testing.T) {
// 	store, err := NewTestStore(t)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	senderId, _ := NewId()
// 	convId, _ := NewId()

// 	msg, err := newMessage(senderId, convId, "hello")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if err = store.CreateMessage(*msg); err != nil {
// 		t.Fatal(err)
// 	}

// 	msgs, err := store.GetMessages(msg.ConvId)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// }

func NewId() (string, error) {
	id, err := uuid.NewV7()
	return id.String(), err
}

func cleanupTestData(session *gocql.Session) error {
	err := session.Query(`
	DROP * FROM messages;
	`).Exec()
	return err

}

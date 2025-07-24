package scylladb

import (
	"errors"
	"fmt"
	"time"

	"github.com/elug3/gochat/pkg/model"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

type MessageStore struct {
	cluster *gocql.ClusterConfig
}

func NewMessageStore(keyspace string, hosts ...string) (*MessageStore, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: "scylla",
		Password: "D9ahbg0wEFO8kxe",
	}
	cluster.PoolConfig.HostSelectionPolicy = gocql.DCAwareRoundRobinPolicy("GCE_US_EAST_1")
	if err := initdb(cluster, keyspace); err != nil {
		return nil, fmt.Errorf("falied to initilization: %w", err)
	}

	store := MessageStore{cluster: cluster}

	return &store, nil
}

func initdb(cluster *gocql.ClusterConfig, keyspace string) error {
	exists, err := ensureKeyspace(cluster, keyspace)
	if err != nil {
		return err
	}
	if !exists {
		session, err := cluster.CreateSession()
		if err != nil {
			return err
		}
		if err = createTable(session); err != nil {
			return err
		}
	}
	return nil
}

// ensureKeyspace ensures that the given keyspace exists in the cluster.
// If it does not exist, the keyspace is created.
// The cluster configuration is then updated to use that keyspace.
func ensureKeyspace(cluster *gocql.ClusterConfig, keyspace string) (exists bool, err error) {
	// Note: The session is temporary; callers must create a session after this call to reflect the new keyspace in queries.
	session, err := cluster.CreateSession()
	if err != nil {
		return false, err
	}
	defer session.Close()

	exists, err = checkKeyspaceExists(session, keyspace)
	if err != nil {
		return false, fmt.Errorf("checkKeyspaceExists: %w", err)
	}
	if !exists {
		if err = createKeyspace(session, keyspace); err != nil {
			return false, fmt.Errorf("createKeyspace: %w", err)
		}
	}
	cluster.Keyspace = keyspace
	return exists, nil
}

// checkKeyspaceExists checks the given keyspace exists.
func checkKeyspaceExists(session *gocql.Session, keyspace string) (bool, error) {

	var result string
	err := session.Query(`
		SELECT keyspace_name
		FROM system_schema.keyspaces
		WHERE keyspace_name = ?;
	`, keyspace).Scan(&result)

	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	if keyspace != result {
		return false, fmt.Errorf("keyspace mismatch: expected %q, got %q", keyspace, result)
	}
	return true, nil
}

func createTable(session *gocql.Session) error {
	err := session.Query(`
	CREATE TABLE messages (
	id UUID,
	conversation_id UUID,
	sender_id UUID,
	content TEXT,
	created_at TIMESTAMP,
	PRIMARY KEY (id, conversation_id)
	);`).Exec()
	return err
}

func createKeyspace(session *gocql.Session, keyspace string) error {
	sq := fmt.Sprintf(`CREATE KEYSPACE %s
	WITH replication = {
		'class': 'SimpleStrategy',
		'replication_factor': 1
	};`, keyspace)

	err := session.Query(sq).Exec()
	return err
}

func (store *MessageStore) GetMessages(convId string) ([]model.Message, error) {
	session, err := store.cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	scanner := session.Query(`
	SELECT id, sender_id, content, created_at
	FROM messages
	WHERE conversation_id = ?
	LIMIT 50
	`, convId).Iter().Scanner()

	msgs := make([]model.Message, 0)
	for scanner.Next() {
		var msg model.Message
		if err = scanner.Scan(&msg.ConvId, &msg.SenderId, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return msgs, nil
}

func (store *MessageStore) CreateMessage(msg model.Message) error {
	session, err := store.cluster.CreateSession()
	if err != nil {
		return err
	}
	err = session.Query(`
	INSERT INTO messages (id, conversation_id, sender_id, content, created_at)
	VALUES (?, ?, ?, ?, ?);
	`, msg.Id, msg.ConvId, msg.SenderId, msg.Content, msg.CreatedAt).Exec()
	if err != nil {
		return err
	}
	return nil
}

func newMessage(senderId, convId string, content string) (*model.Message, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	msg := model.Message{
		Id:        id.String(),
		SenderId:  senderId,
		ConvId:    convId,
		Content:   content,
		CreatedAt: time.Now(),
	}
	return &msg, nil
}

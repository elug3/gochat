package sqlite

import (
	"testing"

	"github.com/elug3/gochat/internal/config"
)

func newTestContactsStore() (*ContactsStore, error) {
	return NewContactsStore(&config.Config{
		NoSave: true,
	})
}

func TestContactsStore_CreateGroup(t *testing.T) {
	store, err := newTestContactsStore()
	if err != nil {
		t.Fatal(err)
	}
	txc, err := store.Begin()
	if err != nil {
		t.Fatal(err)
	}

	group, err := txc.CreateGroup("mygroup")
	if err != nil {
		t.Fatal(err)
	}

	if group.Name != "mygroup" {
		t.Fatalf("unexpected result: want: %q, but got: %q", "mygroup", group.Name)
	}
}

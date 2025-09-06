package sqlite3

import (
	"testing"
)

func newTestContactsStore() (*ContactsStore, error) {
	store, err := NewContactsStore()
	if err != nil {
		return nil, err
	}
	return store, nil
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

package db

import (
	"testing"

	"github.com/sourcenetwork/defradb/document"
)

func newMemoryDB() (*DB, error) {
	opts := &Options{
		Store: "memory",
		Memory: MemoryOptions{
			Size: 1024 * 1000,
		},
	}

	return NewDB(opts)
}

func TestNewDB(t *testing.T) {
	opts := &Options{
		Store: "memory",
		Memory: MemoryOptions{
			Size: 1024 * 1000,
		},
	}

	_, err := NewDB(opts)
	if err != nil {
		t.Error(err)
	}
}

func TestDBSaveSimpleDocument(t *testing.T) {
	db, _ := newMemoryDB()

	testJSONObj := []byte(`{
		"Name": "John",
		"Ages": 21
	}`)

	doc, err := document.NewFromJSON(testJSONObj)
	if err != nil {
		t.Error(err)
		return
	}

	err = db.Save(doc)
	if err != nil {
		t.Error(err)
	}

	db.printDebugDB()
}

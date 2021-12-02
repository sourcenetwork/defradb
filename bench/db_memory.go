// +build memory

package bench

import (
	"fmt"

	badger "github.com/dgraph-io/badger/v3"

	badgerds "github.com/sourcenetwork/defradb/datastores/badger/v3"
	defradb "github.com/sourcenetwork/defradb/db"
)

func newDB() (*defradb.DB, error) {
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	if err != nil {
		return nil, fmt.Errorf("Failed to create badger in-memory store: %w", err)
	}

	return defradb.NewDB(rootstore, nil)
}

func cleanupDB(db *defradb.DB) {
	db.Close()
}

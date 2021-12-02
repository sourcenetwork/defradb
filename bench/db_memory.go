// +build memory

package bench

import (
	"github.com/sourcenetwork/defradb/db"
)

var dbopts = &db.Options{
	Address: "localhost:19181",
	Store:   "memory",
}

func newDB() (*db.DB, error) {
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	if err != nil {
		return nil, fmt.Errorf("Failed to create badger in-memory store: %w", err)
	}

	return defradb.NewDB(rootstore, nil)
}

func cleanupDB(db *db.DB) {
	db.Close()
}

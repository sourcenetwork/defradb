// +build memorymap

package bench

import (
	ds "github.com/ipfs/go-datastore"

	defradb "github.com/sourcenetwork/defradb/db"
)

func newDB() (*defradb.DB, error) {
	rootstore := ds.NewMapDatastore()
	return defradb.NewDB(rootstore, nil)
}

func cleanupDB(db *defradb.DB) {
	db.Close()
}

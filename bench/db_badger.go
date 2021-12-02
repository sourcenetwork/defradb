// +build badger

package bench

import (
	"fmt"
	"io/ioutil"
	"os"

	badger "github.com/dgraph-io/badger/v3"

	badgerds "github.com/sourcenetwork/defradb/datastores/badger/v3"
	defradb "github.com/sourcenetwork/defradb/db"
)

var dbpath string

// handle temp dir in a cross-platform way
func init() {
	dir, err := ioutil.TempDir("", "defra-bench")
	if err != nil {
		panic(err)
	}
	dbpath = dir
}

func newDB() (*defradb.DB, error) {
	rootstore, err := badgerds.NewDatastore(dbpath, badger.DefaultOptions(dbpath))
	if err != nil {
		return nil, fmt.Errorf("Failed to create badger in-memory store: %w", err)
	}

	return defradb.NewDB(rootstore, struct{}{})
}

func cleanupDB(db *defradb.DB) {
	db.Close()
	removeDir(dbopts.Badger.Path)
}

func removeDir(dir string) {
	if err := os.RemoveAll(dir); err != nil {
		fmt.Printf("Error while removing dir: %v\n", err)
	}
}

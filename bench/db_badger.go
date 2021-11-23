// +build badger

package bench

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sourcenetwork/defradb/db"
)

var dbopts = &db.Options{
	Address: "localhost:19181",
	Store:   "badger",
	Badger:  db.BadgerOptions{},
}

// handle temp dir in a cross-platform way
func init() {
	dir, err := ioutil.TempDir("", "defra-bench")
	if err != nil {
		panic(err)
	}
	dbopts.Badger.Path = dir
}

func cleanupDB(db *db.DB) {
	db.Close()
	removeDir(dbopts.Badger.Path)
}

func removeDir(dir string) {
	if err := os.RemoveAll(dir); err != nil {
		fmt.Printf("Error while removing dir: %v\n", err)
	}
}

// +build badger

package bench

import (
	"github.com/sourcenetwork/defradb/db"
)

var dbopts = &db.Options{
	Address: "localhost:19181",
	Store:   "badger",
	Badger: db.BadgerOptions{
		Path: "/tmp/defra-bench",
	},
}

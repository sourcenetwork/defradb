// +build memory

package bench

import (
	"github.com/sourcenetwork/defradb/db"
)

var dbopts = &db.Options{
	Address: "localhost:19181",
	Store:   "memory",
}

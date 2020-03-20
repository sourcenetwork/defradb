package core

import (
	"github.com/sourcenetwork/defradb/store"

	// ds "github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log"
)

// MultiStore is an interface wrapper around the 3 main types of stores needed for
// MerkleCRDTs
type MultiStore interface {
	Data() store.DSReaderWriter
	Head() store.DSReaderWriter
	Dag() *store.DAGStore
	Log() logging.StandardLogger
}

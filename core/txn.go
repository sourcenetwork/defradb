package core

import (
	ds "github.com/ipfs/go-datastore"
)

// Txn is a common interface to the db.Txn struct
type Txn interface {
	ds.Txn
	MultiStore
	Systemstore() DSReaderWriter
}

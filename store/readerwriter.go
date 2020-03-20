package store

import (
	ds "github.com/ipfs/go-datastore"
)

// DSReaderWriter simplifies the interface that is exposed by a
// store.DSReaderWriter into its subcomponents Reader and Writer.
// Using this simplified interface means that both store.DSReaderWriter
// and ds.Txn satisy the interface. Due to go-datastore#113 and
// go-datastore#114 ds.Txn no longer implements store.DSReaderWriter
// Which means we can't swap between the two for Datastores that
// support TxnDatastore.
type DSReaderWriter interface {
	ds.Read
	ds.Write
}

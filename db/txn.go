package db

import (
	"errors"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/store"

	ds "github.com/ipfs/go-datastore"
	ktds "github.com/ipfs/go-datastore/keytransform"
)

var (
	ErrNoTxnSupport = errors.New("The given store has no transaction or batching support")
)

// Txn is a transaction interface for interacting with the Database.
// It carries over the semantics of the underlying datastore regarding
// transactions.
// IE: If the rootstore has full ACID transactions, then so does Txn.
// If the rootstore is a ds.MemoryStore than it'll only have the Batching
// semantics. With no Commit/Discord functionality
type Txn struct {
	ds.Txn

	// wrapped DS
	datastore core.DSReaderWriter // wrapped /data namespace
	headstore core.DSReaderWriter // wrapped /heads namespace
	dagstore  core.DAGStore       // wrapped /blocks namespace
}

// fascade to implement the ds.Txn interface using ds.Batcher
type dummyBatcherTxn struct {
	ds.Read
	ds.Batch
}

func (rb dummyBatcherTxn) Discard() {
	// noop
}

// readonly is only for datastores that support ds.TxnDatastore
func (db *DB) newTxn(readonly bool) (*Txn, error) {
	txn := new(Txn)

	// check if our datastore natively supports transactions or Batching
	txnStore, ok := db.rootstore.(ds.TxnDatastore)
	if ok { // we support transactions
		dstxn, err := txnStore.NewTransaction(readonly)
		if err != nil {
			return nil, err
		}

		txn.Txn = dstxn

	} else if batchStore, ok := db.rootstore.(ds.Batching); ok { // no txn
		batcher, err := batchStore.Batch()
		if err != nil {
			return nil, err
		}

		// hide a ds.Batching store as a ds.Txn
		rb := dummyBatcherTxn{
			Read:  batchStore,
			Batch: batcher,
		}
		txn.Txn = rb
	} else {
		// our datastore implements neither TxnDatastore or Batching
		// for now return error
		return nil, ErrNoTxnSupport
	}

	// add the wrapped datastores using the existing KeyTransform functions from the db
	// @todo Check if KeyTransforms are nil beforehand
	dummyStore := dummyTxnStore{txn.Txn}
	txn.datastore = ktds.Wrap(dummyStore, db.dsKeyTransform)
	txn.headstore = ktds.Wrap(dummyStore, db.dsKeyTransform)
	batchstore := ktds.Wrap(dummyStore, db.dagKeyTransform)
	txn.dagstore = store.NewDAGStore(batchstore)

	return txn, nil
}

// Datastore returns the txn wrapped as a datastore under the /data namespace
func (txn *Txn) Datastore() core.DSReaderWriter {
	return txn.datastore
}

// Headstore returns the txn wrapped as a headstore under the /heads namespace
func (txn *Txn) Headstore() core.DSReaderWriter {
	return txn.headstore
}

// DAGStore returns the txn wrapped as a blockstore for a DAGStore under the /blocks namespace
func (txn *Txn) DAGstore() core.DAGStore {
	return txn.dagstore
}

// Shim to ensure the Txn
type dummyTxnStore struct {
	ds.Txn
}

func (ts dummyTxnStore) Sync(prefix ds.Key) error {
	return ts.Txn.Commit()
}

func (ts dummyTxnStore) Close() error {
	ts.Discard()
	return nil
}

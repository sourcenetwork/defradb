package db

import (
	"errors"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/store"

	ds "github.com/ipfs/go-datastore"
	ktds "github.com/ipfs/go-datastore/keytransform"
)

var (
	// ErrNoTxnSupport occurs when a new transaction is trying to be created from a
	// root datastore that doesn't support ds.TxnDatastore or ds.Batching 8885
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
	datastore core.DSReaderWriter // wrapped txn /data namespace
	headstore core.DSReaderWriter // wrapped txn /heads namespace
	dagstore  core.DAGStore       // wrapped txn /blocks namespace
}

// Txn creates a new transaction which can be set to readonly mode
func (db *DB) NewTxn(readonly bool) (*Txn, error) {
	return db.newTxn(readonly)
}

// readonly is only for datastores that support ds.TxnDatastore
func (db *DB) newTxn(readonly bool) (*Txn, error) {
	db.glock.RLock()
	defer db.glock.RUnlock()
	txn := new(Txn)

	// check if our datastore natively supports transactions or Batching
	txnStore, ok := db.rootstore.(ds.TxnDatastore)
	if ok { // we support transactions
		dstxn, err := txnStore.NewTransaction(readonly)
		if err != nil {
			return nil, err
		}

		txn.Txn = dstxn

	} else if batchStore, ok := db.rootstore.(ds.Batching); ok { // we support Batching
		batcher, err := batchStore.Batch()
		if err != nil {
			return nil, err
		}

		// hide a ds.Batching store as a ds.Txn
		rb := shimBatcherTxn{
			Read:  batchStore,
			Batch: batcher,
		}
		txn.Txn = rb
	} else {
		// our datastore supports neither TxnDatastore or Batching
		// for now return error
		return nil, ErrNoTxnSupport
	}

	// add the wrapped datastores using the existing KeyTransform functions from the db
	// @todo Check if KeyTransforms are nil beforehand
	shimStore := shimTxnStore{txn.Txn}
	txn.datastore = ktds.Wrap(shimStore, db.dsKeyTransform)
	txn.headstore = ktds.Wrap(shimStore, db.hsKeyTransform)
	batchstore := ktds.Wrap(shimStore, db.dagKeyTransform)
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

// DAGstore returns the txn wrapped as a blockstore for a DAGStore under the /blocks namespace
func (txn *Txn) DAGstore() core.DAGStore {
	return txn.dagstore
}

// Shim to make ds.Txn support ds.Datastore
type shimTxnStore struct {
	ds.Txn
}

func (ts shimTxnStore) Sync(prefix ds.Key) error {
	return ts.Txn.Commit()
}

func (ts shimTxnStore) Close() error {
	ts.Discard()
	return nil
}

// shim to make ds.Batch implement ds.Datastore
type shimBatcherTxn struct {
	ds.Read
	ds.Batch
}

func (shimBatcherTxn) Discard() {
	// noop
}

// txn := db.NewTxn()
// users := db.GetCollection("users")
// usersTxn := users.WithTxn(txn)

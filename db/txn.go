// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package db

import (
	"context"
	"errors"

	"github.com/sourcenetwork/defradb/client"
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

// implement interface check
var _ core.Txn = (*Txn)(nil)
var _ client.Txn = (*Txn)(nil)

// Txn is a transaction interface for interacting with the Database.
// It carries over the semantics of the underlying datastore regarding
// transactions.
// IE: If the rootstore has full ACID transactions, then so does Txn.
// If the rootstore is a ds.MemoryStore than it'll only have the Batching
// semantics. With no Commit/Discord functionality
type Txn struct {
	ds.Txn

	// wrapped DS
	systemstore core.DSReaderWriter // wrapped txn /system namespace
	datastore   core.DSReaderWriter // wrapped txn /data namespace
	headstore   core.DSReaderWriter // wrapped txn /heads namespace
	dagstore    core.DAGStore       // wrapped txn /blocks namespace
}

// Txn creates a new transaction which can be set to readonly mode
func (db *DB) NewTxn(ctx context.Context, readonly bool) (*Txn, error) {
	return db.newTxn(ctx, readonly)
}

// readonly is only for datastores that support ds.TxnDatastore
func (db *DB) newTxn(ctx context.Context, readonly bool) (*Txn, error) {
	db.glock.RLock()
	defer db.glock.RUnlock()

	txn := new(Txn)

	// check if our datastore natively supports transactions or Batching
	txnStore, ok := db.rootstore.(ds.TxnDatastore)
	if ok { // we support transactions
		dstxn, err := txnStore.NewTransaction(ctx, readonly)
		if err != nil {
			return nil, err
		}

		txn.Txn = dstxn

		// Note: db.rootstore now has type `ds.Batching`.
	} else {
		batcher, err := db.rootstore.Batch(ctx)
		if err != nil {
			return nil, err
		}

		// hide a ds.Batching store as a ds.Txn
		rb := shimBatcherTxn{
			Read:  db.rootstore,
			Batch: batcher,
		}
		txn.Txn = rb
	}

	// add the wrapped datastores using the existing KeyTransform functions from the db
	// @todo Check if KeyTransforms are nil beforehand
	shimStore := shimTxnStore{txn.Txn}
	txn.systemstore = ktds.Wrap(shimStore, db.ssKeyTransform)
	txn.datastore = ktds.Wrap(shimStore, db.dsKeyTransform)
	txn.headstore = ktds.Wrap(shimStore, db.hsKeyTransform)
	batchstore := ktds.Wrap(shimStore, db.dagKeyTransform)
	txn.dagstore = store.NewDAGStore(batchstore)

	return txn, nil
}

// Systemstore returns the txn wrapped as a systemstore under the /system namespace
func (txn *Txn) Systemstore() core.DSReaderWriter {
	return txn.systemstore
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

func (txn *Txn) IsBatch() bool {
	_, ok := txn.Txn.(shimBatcherTxn)
	return ok
}

// Shim to make ds.Txn support ds.Datastore
type shimTxnStore struct {
	ds.Txn
}

func (ts shimTxnStore) Sync(ctx context.Context, prefix ds.Key) error {
	return ts.Txn.Commit(ctx)
}

func (ts shimTxnStore) Close() error {
	ts.Discard(context.TODO())
	return nil
}

// shim to make ds.Batch implement ds.Datastore
type shimBatcherTxn struct {
	ds.Read
	ds.Batch
}

func (shimBatcherTxn) Discard(_ context.Context) {
	// noop
}

// txn := db.NewTxn()
// users := db.GetCollection("users")
// usersTxn := users.WithTxn(txn)

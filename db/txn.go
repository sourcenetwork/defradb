// Copyright 2022 Democratized Data Foundation
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
	"math/rand"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/store"

	ds "github.com/ipfs/go-datastore"
	ktds "github.com/ipfs/go-datastore/keytransform"
	"github.com/sourcenetwork/defradb/datastores/iterable"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

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
	iterable.IterableTxn
	isBatch bool

	// wrapped DS
	systemstore core.DSReaderWriter // wrapped txn /system namespace
	datastore   core.DSReaderWriter // wrapped txn /data namespace
	headstore   core.DSReaderWriter // wrapped txn /heads namespace
	dagstore    core.DAGStore       // wrapped txn /blocks namespace

	successFns []func()
	errorFns   []func()
}

// NewTxnI returns a new transaction, but using the /client interface
// func (db *DB) NewTxnI(ctx context.Context, readonly bool) (client.Txn, error) {
// 	return db.NewTxn(ctx, readonly)
// }

// Txn creates a new transaction which can be set to readonly mode
func (db *DB) NewTxn(ctx context.Context, readonly bool) (client.Txn, error) {
	return db.newTxn(ctx, readonly)
}

// readonly is only for datastores that support ds.TxnDatastore
func (db *DB) newTxn(ctx context.Context, readonly bool) (*Txn, error) {
	db.glock.RLock()
	defer db.glock.RUnlock()

	txn := new(Txn)

	// check if our datastore natively supports iterable transaction, transactions or batching
	if iterableTxnStore, ok := db.rootstore.(iterable.IterableTxnDatastore); ok {
		dstxn, err := iterableTxnStore.NewIterableTransaction(ctx, readonly)
		if err != nil {
			return nil, err
		}
		txn.IterableTxn = dstxn
	} else if txnStore, ok := db.rootstore.(ds.TxnDatastore); ok {
		dstxn, err := txnStore.NewTransaction(ctx, readonly)
		if err != nil {
			return nil, err
		}
		txn.IterableTxn = iterable.NewIterableTransaction(dstxn)
		// Note: db.rootstore now has type `ds.Batching`.
	} else {
		batcher, err := db.rootstore.Batch(ctx)
		if err != nil {
			return nil, err
		}

		// hide a ds.Batching store as a ds.Txn
		rb := store.ShimBatcherTxn{
			Read:  db.rootstore,
			Batch: batcher,
		}
		txn.IterableTxn = iterable.NewIterableTransaction(rb)
		txn.isBatch = true
	}

	// add the wrapped datastores using the existing KeyTransform functions from the db
	// @todo Check if KeyTransforms are nil beforehand

	// debug stuff... ignore
	//
	// txnid := RandStringRunes(5)
	// txn.systemstore = ds.NewLogDatastore(ktds.Wrap(shimStore, db.ssKeyTransform), fmt.Sprintf("%s:systemstore", txnid))
	// txn.datastore = ds.NewLogDatastore(ktds.Wrap(shimStore, db.dsKeyTransform), fmt.Sprintf("%s:datastore", txnid))
	// txn.headstore = ds.NewLogDatastore(ktds.Wrap(shimStore, db.hsKeyTransform), fmt.Sprintf("%s:headstore", txnid))
	// batchstore := ds.NewLogDatastore(ktds.Wrap(shimStore, db.dagKeyTransform), fmt.Sprintf("%s:dagstore", txnid))

	shimStore := store.ShimTxnStore{Txn: txn.IterableTxn}
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

// Rootstore returns the underlying txn as a DSReaderWriter to implement
// the MultiStore interface
func (txn *Txn) Rootstore() core.DSReaderWriter {
	return txn.IterableTxn
}

func (txn *Txn) IsBatch() bool {
	return txn.isBatch
}

func (txn *Txn) Commit(ctx context.Context) error {
	if err := txn.IterableTxn.Commit(ctx); err != nil {
		txn.runErrorFns(ctx)
		return err
	}
	txn.runSuccessFns(ctx)
	return nil
}

func (txn *Txn) OnSuccess(fn func()) {
	if fn == nil {
		return
	}
	txn.successFns = append(txn.successFns, fn)
}

func (txn *Txn) OnError(fn func()) {
	if fn == nil {
		return
	}
	txn.errorFns = append(txn.errorFns, fn)
}

func (txn *Txn) runErrorFns(ctx context.Context) {
	for _, fn := range txn.errorFns {
		fn()
	}
}

func (txn *Txn) runSuccessFns(ctx context.Context) {
	for _, fn := range txn.successFns {
		fn()
	}
}

// txn := db.NewTxn()
// users := db.GetCollection("users")
// usersTxn := users.WithTxn(txn)

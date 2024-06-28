// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package datastore

import (
	"context"
	"sync"

	ds "github.com/ipfs/go-datastore"
)

type concurrentTxn struct {
	ds.Txn

	// Some datastore don't support concurrent operation within a single transaction. `concurrentTxn` with its
	// mutex enable those concurrent operations. This was implemented for DefraDB's DAG sync process.
	// Since the DAG sync process is highly concurrent and has been made to operate on a single transaction
	// to eliminate the potential for deadlock (DAG being left in an incomplete state without a way to obviously
	// detect it), we need to add a mutex to ensure thread safety.
	mu sync.Mutex
}

// NewConcurrentTxnFrom creates a new Txn from rootstore that supports concurrent API calls
func NewConcurrentTxnFrom(ctx context.Context, rootstore ds.TxnDatastore, id uint64, readonly bool) (Txn, error) {
	rootTxn, err := newTxnFrom(ctx, rootstore, readonly)
	if err != nil {
		return nil, err
	}
	rootConcurentTxn := &concurrentTxn{Txn: rootTxn}
	multistore := MultiStoreFrom(rootConcurentTxn)
	return &txn{
		t:          rootConcurentTxn,
		MultiStore: multistore,
		id:         id,
	}, nil
}

// Delete implements ds.Delete
func (t *concurrentTxn) Delete(ctx context.Context, key ds.Key) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Txn.Delete(ctx, key)
}

// Get implements ds.Get
func (t *concurrentTxn) Get(ctx context.Context, key ds.Key) ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Txn.Get(ctx, key)
}

// Has implements ds.Has
func (t *concurrentTxn) Has(ctx context.Context, key ds.Key) (bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Txn.Has(ctx, key)
}

// Put implements ds.Put
func (t *concurrentTxn) Put(ctx context.Context, key ds.Key, value []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Txn.Put(ctx, key, value)
}

// Sync executes the transaction.
func (t *concurrentTxn) Sync(ctx context.Context, prefix ds.Key) error {
	return t.Commit(ctx)
}

// Close discards the transaction.
func (t *concurrentTxn) Close() error {
	t.Discard(context.TODO())
	return nil
}

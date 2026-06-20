// Copyright 2026 Democratized Data Foundation
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
	"time"

	"github.com/sourcenetwork/defradb/internal/datastore"
)

// lockToken is a rootstore-backed pseudo-transaction used to hold a collection write lock
// without opening an underlying store transaction.
//
// It exists because corekv-leveldb permits only one open transaction at a time and blocks all
// non-transactional writes while one is open (see https://github.com/sourcenetwork/defradb/issues/4959).
// Operations such as Truncate and RefreshView must perform non-transactional (txn-free) writes, so
// they cannot hold a store transaction open for the duration. The lock token lets them hold the
// collection write lock - which is what actually provides reader isolation - while every read and
// write goes directly to the rootstore with no transaction open.
//
// It satisfies datastore.Txn (the store accessors are backed by a Multistore over the rootstore) so
// that read paths which do datastore.CtxMustGetTxn(ctx).Systemstore() continue to work, and it
// carries a unique id because the cache layer keys per-transaction caches by id. Its lifecycle is
// purely the registered close callbacks: releasing it frees the collection lock(s) acquired under it
// and disposes any per-token caches.
type lockToken struct {
	*datastore.Multistore

	id uint64
	ts time.Time

	successFns []func()
	errorFns   []func()
	discardFns []func()

	successAsyncFns []func()
	errorAsyncFns   []func()
	discardAsyncFns []func()

	closed bool
}

var _ datastore.Txn = (*lockToken)(nil)

// newLockToken returns a lock token backed by the rootstore, with a unique id drawn from the same
// allocator as real transactions so it never collides with one.
func (db *DB) newLockToken() *lockToken {
	return &lockToken{
		Multistore: datastore.NewMultistore(db.rootstore, db.lockSet, db.blockStoreChunkSize),
		id:         db.previousTxnID.Add(1),
		ts:         time.Now(),
	}
}

func (t *lockToken) ID() uint64         { return t.id }
func (t *lockToken) StartTS() time.Time { return t.ts }

func (t *lockToken) OnSuccess(fn func()) { t.successFns = append(t.successFns, fn) }
func (t *lockToken) OnError(fn func())   { t.errorFns = append(t.errorFns, fn) }
func (t *lockToken) OnDiscard(fn func()) { t.discardFns = append(t.discardFns, fn) }

func (t *lockToken) OnSuccessAsync(fn func()) { t.successAsyncFns = append(t.successAsyncFns, fn) }
func (t *lockToken) OnErrorAsync(fn func())   { t.errorAsyncFns = append(t.errorAsyncFns, fn) }
func (t *lockToken) OnDiscardAsync(fn func()) { t.discardAsyncFns = append(t.discardAsyncFns, fn) }

// release fires the registered close callbacks exactly once, freeing the collection lock(s) held
// under this token and disposing any per-token caches.
//
// All current consumers (the lock set and the cache layer) register identical callbacks on success,
// error, and discard, and the cache layer writes through to the store before caching, so firing a
// single set here is correct. Repeated calls are no-ops.
func (t *lockToken) release() {
	if t.closed {
		return
	}
	t.closed = true

	for _, fn := range t.discardAsyncFns {
		go fn()
	}
	for _, fn := range t.discardFns {
		fn()
	}
}

// Commit releases the token. There is nothing to commit - the token never opens a store transaction
// and all writes made under it are already durable.
func (t *lockToken) Commit() error {
	t.release()
	return nil
}

// Discard releases the token.
func (t *lockToken) Discard() {
	t.release()
}

// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cache

import (
	"context"
	"sync"

	"github.com/sourcenetwork/defradb/internal/datastore"
)

// TxnRepository is responsible for managing store access at the `transaction`
// layer as detailed in the README.md
//
// On read, it will first attempt to fetch the requested value from the transaction cache,
// only after failing that will it attempt to read from store.
type TxnRepository[T any, TV any] struct {
	// newCache creates a new instance of `Cache[T, TV]`
	newCache func() Cache[T, TV]

	// txnLock protects `cache` from concurrent read-writes
	txnLock sync.RWMutex
	// cache is the set of transaction specific caches keyed by transaction id.
	cache map[uint64]Cache[T, TV]

	// lower is the next layer down in the repository stack.
	//
	// For now, this is the `store` layer, but in the future this will likely be a global cache.
	lower Repository[T, TV]
}

var _ Repository[any, any] = (*TxnRepository[any, any])(nil)

func NewTxnLayer[T any, TV any](
	newCache func() Cache[T, TV],
	lower Repository[T, TV],
) *TxnRepository[T, TV] {
	return &TxnRepository[T, TV]{
		newCache: newCache,
		lower:    lower,
		cache:    map[uint64]Cache[T, TV]{},
	}
}

// registerTxn returns the transaction's cached repository.
//
// If a repository for this transaction does not exist, it will create it. The created
// repositories will be disposed of upon transaction discard, error, or success.  The
// transaction repository may be recreated after initial disposal as many times as required.
func (l *TxnRepository[T, TV]) registerTxn(ctx context.Context) Cache[T, TV] {
	txn := datastore.CtxMustGetTxn(ctx)
	id := txn.ID()

	l.txnLock.Lock()
	txnCache, ok := l.cache[id]
	if ok {
		l.txnLock.Unlock()
		return txnCache
	}

	txnCache = l.newCache()
	l.cache[id] = txnCache
	l.txnLock.Unlock()

	// todo - an integration test for this is required, and make sure at least one tests discard and reuse:
	// https://github.com/sourcenetwork/defradb/issues/4268
	onTxnClose := func() {
		l.txnLock.Lock()
		_, ok := l.cache[id]
		if !ok {
			l.txnLock.Unlock()
			return
		}
		delete(l.cache, id)
		l.txnLock.Unlock()
	}

	txn.OnDiscard(onTxnClose)
	txn.OnError(onTxnClose)
	txn.OnSuccess(onTxnClose)

	return txnCache
}

func (l *TxnRepository[T, TV]) TryGet(ctx context.Context, key T) (TV, bool, error) {
	txnCache := l.registerTxn(ctx)

	if value, ok := txnCache.TryGet(key); ok {
		return value, true, nil
	}

	value, ok, err := l.lower.TryGet(ctx, key)
	if err != nil || !ok {
		var d TV
		return d, false, err
	}

	txnCache.Cache(value)

	return value, ok, err
}

func (l *TxnRepository[T, TV]) Write(ctx context.Context, value TV) error {
	// Always write to the lower cache *first*, making sure any errors are surfaced from the store
	// before polluting higher level caches
	err := l.lower.Write(ctx, value)
	if err != nil {
		return err
	}

	txnCache := l.registerTxn(ctx)
	txnCache.Cache(value)

	return nil
}

func (l *TxnRepository[T, TV]) Delete(ctx context.Context, key T) error {
	// Always delete from the lower cache *first*, making sure any errors are surfaced from the store
	// before polluting higher level caches
	err := l.lower.Delete(ctx, key)
	if err != nil {
		return err
	}

	txnCache := l.registerTxn(ctx)
	txnCache.Remove(key)

	return nil
}

func (l *TxnRepository[T, TV]) Forbid(value TV) {
	l.lower.Forbid(value)

	l.txnLock.RLock()
	for _, cache := range l.cache {
		cache.Forbid(value)
	}
	l.txnLock.RUnlock()
}

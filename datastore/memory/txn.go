// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package memory

import (
	"context"
	"sync"
	"sync/atomic"

	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	"github.com/tidwall/btree"
)

// basicTxn implements ds.Txn
type basicTxn struct {
	ops *btree.BTreeG[dsItem]
	ds  *Datastore
	// Version of the datastore when the transaction was initiated.
	dsVersion *uint64
	readOnly  bool
	discarded bool

	closed  bool
	closeLk sync.RWMutex
}

var _ ds.Txn = (*basicTxn)(nil)

func (t *basicTxn) getDSVersion() uint64 {
	return atomic.LoadUint64(t.dsVersion)
}

func (t *basicTxn) getTxnVersion() uint64 {
	return t.getDSVersion() + 1
}

// Delete implements ds.Delete.
func (t *basicTxn) Delete(ctx context.Context, key ds.Key) error {
	t.closeLk.RLock()
	defer t.closeLk.RUnlock()
	if t.closed {
		return ErrClosed
	}

	if t.discarded {
		return ErrTxnDiscarded
	}
	if t.readOnly {
		return ErrReadOnlyTxn
	}

	item := t.get(ctx, key)
	if item.key == "" || item.isDeleted {
		// if the key doesn't exist of the item is already deleted, this is a no-op.
		return nil
	}

	t.ops.Set(dsItem{key: key.String(), version: t.getTxnVersion(), isDeleted: true})
	return nil
}

func (t *basicTxn) get(ctx context.Context, key ds.Key) dsItem {
	result := dsItem{}
	t.ops.Descend(dsItem{key: key.String(), version: t.getTxnVersion()}, func(item dsItem) bool {
		if key.String() == item.key {
			result = item
		}
		// We only care about the last version so we stop iterating right away by returning false.
		return false
	})
	if result.key == "" {
		result = t.ds.get(ctx, key, t.getDSVersion())
		result.isGet = true
		if result.key == "" {
			// If the datastore doesn't have the item, we still need to track it
			// to check for merge conflicts.
			result.key = key.String()
		}
		t.ops.Set(result)
	}
	return result
}

// Get implements ds.Get.
func (t *basicTxn) Get(ctx context.Context, key ds.Key) ([]byte, error) {
	t.closeLk.RLock()
	defer t.closeLk.RUnlock()
	if t.closed {
		return nil, ErrClosed
	}

	if t.discarded {
		return nil, ErrTxnDiscarded
	}
	result := t.get(ctx, key)
	if result.version == 0 || result.isDeleted {
		return nil, ds.ErrNotFound
	}
	return result.val, nil
}

// GetSize implements ds.GetSize.
func (t *basicTxn) GetSize(ctx context.Context, key ds.Key) (size int, err error) {
	t.closeLk.RLock()
	defer t.closeLk.RUnlock()
	if t.closed {
		return 0, ErrClosed
	}

	if t.discarded {
		return 0, ErrTxnDiscarded
	}
	result := t.get(ctx, key)
	if result.version == 0 || result.isDeleted {
		return 0, ds.ErrNotFound
	}
	return len(result.val), nil
}

// Has implements ds.Has.
func (t *basicTxn) Has(ctx context.Context, key ds.Key) (exists bool, err error) {
	t.closeLk.RLock()
	defer t.closeLk.RUnlock()
	if t.closed {
		return false, ErrClosed
	}

	if t.discarded {
		return false, ErrTxnDiscarded
	}
	result := t.get(ctx, key)
	if result.version == 0 || result.isDeleted {
		return false, nil
	}
	return true, nil
}

// Put implements ds.Put.
func (t *basicTxn) Put(ctx context.Context, key ds.Key, value []byte) error {
	t.closeLk.RLock()
	defer t.closeLk.RUnlock()
	if t.closed {
		return ErrClosed
	}

	if t.discarded {
		return ErrTxnDiscarded
	}
	if t.readOnly {
		return ErrReadOnlyTxn
	}
	t.ops.Set(dsItem{key: key.String(), version: t.getTxnVersion(), val: value})

	return nil
}

// Query implements ds.Query.
func (t *basicTxn) Query(ctx context.Context, q dsq.Query) (dsq.Results, error) {
	t.closeLk.RLock()
	defer t.closeLk.RUnlock()
	if t.closed {
		return nil, ErrClosed
	}

	if t.discarded {
		return nil, ErrTxnDiscarded
	}
	// best effort allocation
	re := make([]dsq.Entry, 0, t.ds.values.Height()+t.ops.Height())
	iter := t.ds.values.Iter()
	iterOps := t.ops.Iter()
	iterOpsHasValue := iterOps.Next()
	// iterate over the underlying store and ensure that ops with keys smaller than or equal to
	// the key of the underlying store are added with priority.
	for iter.Next() {
		// fast forward to last inserted version
		item := iter.Item()
		for iter.Next() {
			if item.key == iter.Item().key {
				item = iter.Item()
				continue
			}
			iter.Prev()
			break
		}

		// handle all ops that come before the current item's key or equal to the current item's key
		for iterOpsHasValue && iterOps.Item().key <= item.key {
			if iterOps.Item().key == item.key {
				item = iterOps.Item()
			} else if !iterOps.Item().isDeleted && !iterOps.Item().isGet {
				re = append(re, setEntry(iterOps.Item().key, iterOps.Item().val, q))
			}
			iterOpsHasValue = iterOps.Next()
		}

		if item.isDeleted {
			continue
		}

		re = append(re, setEntry(item.key, item.val, q))
	}

	iter.Release()

	// add the remaining ops
	for iterOpsHasValue {
		if !iterOps.Item().isDeleted && !iterOps.Item().isGet {
			re = append(re, setEntry(iterOps.Item().key, iterOps.Item().val, q))
		}
		iterOpsHasValue = iterOps.Next()
	}

	iterOps.Release()

	r := dsq.ResultsWithEntries(q, re)
	r = dsq.NaiveQueryApply(q, r)
	return r, nil
}

func setEntry(key string, value []byte, q dsq.Query) dsq.Entry {
	e := dsq.Entry{
		Key:  key,
		Size: len(value),
	}
	if !q.KeysOnly {
		e.Value = value
	}
	return e
}

// Discard removes all the operations added to the transaction.
func (t *basicTxn) Discard(ctx context.Context) {
	if t.discarded {
		return
	}
	t.ops.Clear()
	t.clearInFlightTxn(ctx)
	t.discarded = true
}

// Commit saves the operations to the underlying datastore.
func (t *basicTxn) Commit(ctx context.Context) error {
	t.closeLk.RLock()
	defer t.closeLk.RUnlock()
	if t.closed {
		return ErrClosed
	}

	if t.discarded {
		return ErrTxnDiscarded
	}
	defer t.Discard(ctx)

	if !t.readOnly {
		return t.ds.commit(ctx, t)
	}

	return nil
}

func (t *basicTxn) checkForConflicts(ctx context.Context) error {
	if t.getDSVersion() == t.ds.getVersion() {
		return nil
	}
	iter := t.ops.Iter()
	defer iter.Release()
	for iter.Next() {
		item := iter.Item()
		if !item.isGet {
			// Conflict should only occur if an item has been updated
			// after we've read it within the transaction.
			continue
		}
		expectedItem := t.ds.get(ctx, ds.NewKey(item.key), t.getDSVersion())
		latestItem := t.ds.get(ctx, ds.NewKey(item.key), t.ds.getVersion())
		if latestItem.version != expectedItem.version {
			return ErrTxnConflict
		}
	}
	return nil
}

func (t *basicTxn) clearInFlightTxn(ctx context.Context) {
	t.ds.inFlightTxn.Delete(
		dsTxn{
			dsVersion:  t.getDSVersion(),
			txnVersion: t.getTxnVersion(),
		},
	)
	t.ds.clearOldInFlightTxn(ctx)
}

func (t *basicTxn) close() {
	t.closeLk.Lock()
	defer t.closeLk.Unlock()
	t.closed = true
}

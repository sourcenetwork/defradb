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
	"sync/atomic"

	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	"github.com/tidwall/btree"
)

// basicTxn implements ds.Txn
type basicTxn struct {
	ops       *btree.BTreeG[item]
	ds        *Datastore
	version   *uint64
	readOnly  bool
	discarded bool
}

var _ ds.Txn = (*basicTxn)(nil)

func (t *basicTxn) getVersion() uint64 {
	return atomic.LoadUint64(t.version)
}

func (t *basicTxn) get(ctx context.Context, key ds.Key, version uint64) item {
	result := item{}
	// We only care about the last version so we stop iterating right away by returning false.
	t.ops.Descend(item{key: key.String(), version: version}, func(item item) bool {
		if key.String() == item.key {
			result = item
		}
		return false
	})
	if result.key == "" {
		result = t.ds.get(ctx, key, t.getVersion())
	}
	return result
}

// Get implements ds.Get
func (t *basicTxn) Get(ctx context.Context, key ds.Key) ([]byte, error) {
	if t.discarded {
		return nil, ErrTxnDiscarded
	}
	result := t.get(ctx, key, t.getVersion())
	if result.key == "" || result.isDeleted {
		return nil, ds.ErrNotFound
	}
	return result.val, nil
}

// GetSize implements ds.GetSize
func (t *basicTxn) GetSize(ctx context.Context, key ds.Key) (size int, err error) {
	if t.discarded {
		return 0, ErrTxnDiscarded
	}
	result := t.get(ctx, key, t.getVersion())
	if result.key == "" || result.isDeleted {
		return 0, ds.ErrNotFound
	}
	return len(result.val), nil
}

// Has implements ds.Has
func (t *basicTxn) Has(ctx context.Context, key ds.Key) (exists bool, err error) {
	if t.discarded {
		return false, ErrTxnDiscarded
	}
	result := t.get(ctx, key, t.getVersion())
	if result.key == "" || result.isDeleted {
		return false, nil
	}
	return true, nil
}

// Put implements ds.Put
func (t *basicTxn) Put(ctx context.Context, key ds.Key, value []byte) error {
	if t.discarded {
		return ErrTxnDiscarded
	}
	if t.readOnly {
		return ErrReadOnlyTxn
	}
	t.ops.Set(item{key: key.String(), version: t.getVersion(), val: value})

	return nil
}

// Query implements ds.Query
func (t *basicTxn) Query(ctx context.Context, q dsq.Query) (dsq.Results, error) {
	if t.discarded {
		return nil, ErrTxnDiscarded
	}
	// best effort allocation
	re := make([]dsq.Entry, 0, t.ds.values.Len()+t.ops.Len())
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
			} else if !iterOps.Item().isDeleted {
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
		if !iterOps.Item().isDeleted {
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

// Delete implements ds.Delete
func (t *basicTxn) Delete(ctx context.Context, key ds.Key) error {
	if t.discarded {
		return ErrTxnDiscarded
	}
	if t.readOnly {
		return ErrReadOnlyTxn
	}

	t.ops.Set(item{key: key.String(), version: t.getVersion(), isDeleted: true})
	return nil
}

// Discard removes all the operations added to the transaction
func (t *basicTxn) Discard(ctx context.Context) {
	if t.discarded {
		return
	}
	t.ops.Clear()
	t.discarded = true
}

// Commit saves the operations to the underlying datastore
func (t *basicTxn) Commit(ctx context.Context) error {
	if t.discarded {
		return ErrTxnDiscarded
	}
	if t.readOnly {
		return ErrReadOnlyTxn
	}

	iter := t.ops.Iter()
	for iter.Next() {
		t.ds.values.Set(iter.Item())
	}

	iter.Release()

	t.Discard(ctx)

	return nil
}

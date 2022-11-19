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

	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	"github.com/tidwall/btree"
)

// basicTxn implements ds.Txn
type basicTxn struct {
	ops      *btree.Map[string, op]
	ds       *Datastore
	readOnly bool
}

var _ ds.Txn = (*basicTxn)(nil)

// newTransaction returns a ds.Txn datastore
func newTransaction(d *Datastore, readOnly bool) ds.Txn {
	return &basicTxn{
		ops:      btree.NewMap[string, op](2),
		ds:       d,
		readOnly: readOnly,
	}
}

// Get implements ds.Get
func (t *basicTxn) Get(ctx context.Context, key ds.Key) ([]byte, error) {
	if op, exists := t.ops.Get(key.String()); exists {
		if op.delete {
			return nil, ds.ErrNotFound
		}
		return op.value, nil
	}
	return t.ds.Get(ctx, key)
}

// GetSize implements ds.GetSize
func (t *basicTxn) GetSize(ctx context.Context, key ds.Key) (size int, err error) {
	if op, ok := t.ops.Get(key.String()); ok {
		if op.delete {
			return 0, ds.ErrNotFound
		}
		return len(op.value), nil
	}
	return t.ds.GetSize(ctx, key)
}

// Has implements ds.Has
func (t *basicTxn) Has(ctx context.Context, key ds.Key) (exists bool, err error) {
	if op, ok := t.ops.Get(key.String()); ok {
		if op.delete {
			return false, nil
		}
		return true, nil
	}
	return t.ds.Has(ctx, key)
}

// Put implements ds.Put
func (t *basicTxn) Put(ctx context.Context, key ds.Key, value []byte) error {
	if t.readOnly {
		return ErrReadOnlyTxn
	}

	t.ops.Set(key.String(), op{value: value})
	return nil
}

// Query implements ds.Query
func (t *basicTxn) Query(ctx context.Context, q dsq.Query) (dsq.Results, error) {
	// best effort allocation
	re := make([]dsq.Entry, 0, t.ds.values.Len()+t.ops.Len())
	iter := t.ds.values.Iter()
	iterOps := t.ops.Iter()
	iterOpsHasMore := iterOps.Next()
	for iter.Next() {
		for {
			if !iterOpsHasMore || iterOps.Key() > iter.Key() {
				break
			}
			if iterOps.Value().delete {
				iterOpsHasMore = iterOps.Next()
				iter.Next()
				continue
			}
			e := dsq.Entry{
				Key:  iterOps.Key(),
				Size: len(iterOps.Value().value),
			}
			if !q.KeysOnly {
				e.Value = iterOps.Value().value
			}
			re = append(re, e)
			iterOpsHasMore = iterOps.Next()
			iter.Next()
			continue
		}
		e := dsq.Entry{
			Key:  iter.Key(),
			Size: len(iter.Value()),
		}
		if !q.KeysOnly {
			e.Value = iter.Value()
		}
		re = append(re, e)
	}

	for {
		if !iterOpsHasMore {
			break
		}
		if iterOps.Value().delete {
			iterOpsHasMore = iterOps.Next()
			continue
		}
		e := dsq.Entry{
			Key:  iterOps.Key(),
			Size: len(iterOps.Value().value),
		}
		if !q.KeysOnly {
			e.Value = iterOps.Value().value
		}
		re = append(re, e)
		iterOpsHasMore = iterOps.Next()
	}

	r := dsq.ResultsWithEntries(q, re)
	r = dsq.NaiveQueryApply(q, r)
	return r, nil
}

// Delete implements ds.Delete
func (t *basicTxn) Delete(ctx context.Context, key ds.Key) error {
	if t.readOnly {
		return ErrReadOnlyTxn
	}

	t.ops.Set(key.String(), op{delete: true})
	return nil
}

// Discard removes all the operations added to the transaction
func (t *basicTxn) Discard(ctx context.Context) {
	t.ops.Clear()
}

// Commit saves the operations to the underlying datastore
func (t *basicTxn) Commit(ctx context.Context) error {
	if t.readOnly {
		return ErrReadOnlyTxn
	}

	t.ds.txnmu.Lock()
	defer t.ds.txnmu.Unlock()
	iter := t.ops.Iter()
	for iter.Next() {
		if iter.Value().delete {
			t.ds.values.Delete(iter.Key())
		} else {
			t.ds.values.Set(iter.Key(), iter.Value().value)
		}
	}

	return nil
}

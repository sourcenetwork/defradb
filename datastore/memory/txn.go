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

	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
)

// basicTxn implements ds.Txn
type basicTxn struct {
	syncLock sync.Mutex
	ops      map[ds.Key]op
	target   *Store
	readOnly bool
}

var _ ds.Txn = (*basicTxn)(nil)

// NewTransaction returns a ds.Txn datastore
func NewTransaction(d *Store, readOnly bool) ds.Txn {
	return &basicTxn{
		ops:      make(map[ds.Key]op),
		target:   d,
		readOnly: readOnly,
	}
}

// Get implements ds.Get
func (t *basicTxn) Get(ctx context.Context, key ds.Key) ([]byte, error) {
	t.syncLock.Lock()
	defer t.syncLock.Unlock()

	if op, ok := t.ops[key]; ok {
		if op.delete {
			return nil, ds.ErrNotFound
		}
		return op.value, nil
	}
	return t.target.Get(ctx, key)
}

// GetSize implements ds.GetSize
func (t *basicTxn) GetSize(ctx context.Context, key ds.Key) (size int, err error) {
	t.syncLock.Lock()
	defer t.syncLock.Unlock()

	if op, ok := t.ops[key]; ok {
		if op.delete {
			return -1, ds.ErrNotFound
		}
		return len(op.value), nil
	}
	return t.target.GetSize(ctx, key)
}

// Has implements ds.Has
func (t *basicTxn) Has(ctx context.Context, key ds.Key) (exists bool, err error) {
	t.syncLock.Lock()
	defer t.syncLock.Unlock()

	if op, ok := t.ops[key]; ok {
		if op.delete {
			return false, nil
		}
		return true, nil
	}
	return t.target.Has(ctx, key)
}

// Put implements ds.Put
func (t *basicTxn) Put(ctx context.Context, key ds.Key, value []byte) error {
	if t.readOnly {
		return nil
	}

	t.syncLock.Lock()
	defer t.syncLock.Unlock()

	t.ops[key] = op{value: value}
	return nil
}

// Query implements ds.Query
func (t *basicTxn) Query(ctx context.Context, q dsq.Query) (dsq.Results, error) {
	t.syncLock.Lock()
	defer t.syncLock.Unlock()
	t.target.syncLock.Lock()
	defer t.target.syncLock.Unlock()

	// best effort allocation
	re := make([]dsq.Entry, 0, len(t.target.values)+len(t.ops))
	handledOps := make(map[ds.Key]struct{})
	for k, v := range t.target.values {
		e := dsq.Entry{}
		if op, exists := t.ops[k]; exists {
			handledOps[k] = struct{}{}
			if op.delete {
				continue
			}
			e.Key = k.String()
			e.Size = len(op.value)
			if !q.KeysOnly {
				e.Value = op.value
			}
		} else {
			e.Key = k.String()
			e.Size = len(v)
			if !q.KeysOnly {
				e.Value = v
			}
		}

		re = append(re, e)
	}

	for k, v := range t.ops {
		if _, handled := handledOps[k]; handled {
			continue
		}

		if v.delete {
			continue
		}
		e := dsq.Entry{Key: k.String(), Size: len(v.value)}
		if !q.KeysOnly {
			e.Value = v.value
		}
		re = append(re, e)
	}

	r := dsq.ResultsWithEntries(q, re)
	r = dsq.NaiveQueryApply(q, r)
	return r, nil
}

// Delete implements ds.Delete
func (t *basicTxn) Delete(ctx context.Context, key ds.Key) error {
	if t.readOnly {
		return nil
	}

	t.syncLock.Lock()
	defer t.syncLock.Unlock()

	t.ops[key] = op{delete: true}
	return nil
}

// Discard removes all the operations added to the transaction
func (t *basicTxn) Discard(ctx context.Context) {
	t.syncLock.Lock()
	defer t.syncLock.Unlock()

	t.ops = make(map[ds.Key]op)
}

// Commit saves the operations to the target datastore
func (t *basicTxn) Commit(ctx context.Context) error {
	if t.readOnly {
		return nil
	}

	t.syncLock.Lock()
	defer t.syncLock.Unlock()
	t.target.syncLock.Lock()
	defer t.target.syncLock.Unlock()

	for k, op := range t.ops {
		if op.delete {
			delete(t.target.values, k)
		} else {
			t.target.values[k] = op.value
		}
	}

	return nil
}

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
	"github.com/tidwall/btree"
)

// Store uses a btree for internal storage.
type Store struct {
	mu     sync.Mutex
	values *btree.Map[string, []byte]
}

var _ ds.Datastore = (*Store)(nil)
var _ ds.Batching = (*Store)(nil)
var _ ds.TxnFeature = (*Store)(nil)

// NewStore constructs an empty Store.
func NewStore() (d *Store) {
	return &Store{
		values: btree.NewMap[string, []byte](2),
	}
}

// Batch return a ds.Batch datastore based on Store
func (d *Store) Batch(ctx context.Context) (ds.Batch, error) {
	return NewBasicBatch(d), nil
}

func (d *Store) Close() error {
	return nil
}

// Delete implements ds.Delete
func (d *Store) Delete(ctx context.Context, key ds.Key) (err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.values.Delete(key.String())
	return nil
}

// Get implements ds.Get
func (d *Store) Get(ctx context.Context, key ds.Key) (value []byte, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if val, exists := d.values.Get(key.String()); exists {
		return val, nil
	}
	return nil, ds.ErrNotFound
}

// GetSize implements ds.GetSize
func (d *Store) GetSize(ctx context.Context, key ds.Key) (size int, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if val, exists := d.values.Get(key.String()); exists {
		return len(val), nil
	}
	return -1, ds.ErrNotFound
}

// Has implements ds.Has
func (d *Store) Has(ctx context.Context, key ds.Key) (exists bool, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, exists = d.values.Get(key.String())
	return exists, nil
}

// NewTransaction return a ds.Txn datastore based on Store
func (d *Store) NewTransaction(ctx context.Context, readOnly bool) (ds.Txn, error) {
	return NewTransaction(d, readOnly), nil
}

// Put implements ds.Put
func (d *Store) Put(ctx context.Context, key ds.Key, value []byte) (err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.values.Set(key.String(), value)
	return nil
}

// Query implements ds.Query
func (d *Store) Query(ctx context.Context, q dsq.Query) (dsq.Results, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	re := make([]dsq.Entry, 0, d.values.Len())
	iter := d.values.Iter()
	for iter.Next() {
		e := dsq.Entry{Key: iter.Key(), Size: len(iter.Value())}
		if !q.KeysOnly {
			e.Value = iter.Value()
		}
		re = append(re, e)
	}
	r := dsq.ResultsWithEntries(q, re)
	r = dsq.NaiveQueryApply(q, r)
	return r, nil
}

// Sync implements ds.Sync
func (d *Store) Sync(ctx context.Context, prefix ds.Key) error {
	return nil
}

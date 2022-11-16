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

// Store uses a standard Go map for internal storage.
type Store struct {
	syncLock sync.Mutex
	values   map[ds.Key][]byte
}

var _ ds.Datastore = (*Store)(nil)
var _ ds.Batching = (*Store)(nil)
var _ ds.TxnFeature = (*Store)(nil)

// NewStore constructs a Store. It is _not_ thread-safe by
// default, wrap using sync.MutexWrap if you need thread safety (the answer here
// is usually yes).
func NewStore() (d *Store) {
	return &Store{
		values: make(map[ds.Key][]byte),
	}
}

func (d *Store) Batch(ctx context.Context) (ds.Batch, error) {
	return NewBasicBatch(d), nil
}

func (d *Store) Close() error {
	return nil
}

// Delete implements Datastore.Delete
func (d *Store) Delete(ctx context.Context, key ds.Key) (err error) {
	d.syncLock.Lock()
	defer d.syncLock.Unlock()

	delete(d.values, key)
	return nil
}

// Get implements Datastore.Get
func (d *Store) Get(ctx context.Context, key ds.Key) (value []byte, err error) {
	d.syncLock.Lock()
	defer d.syncLock.Unlock()

	val, found := d.values[key]
	if !found {
		return nil, ds.ErrNotFound
	}
	return val, nil
}

// GetSize implements Datastore.GetSize
func (d *Store) GetSize(ctx context.Context, key ds.Key) (size int, err error) {
	d.syncLock.Lock()
	defer d.syncLock.Unlock()

	if v, found := d.values[key]; found {
		return len(v), nil
	}
	return -1, ds.ErrNotFound
}

// Has implements Datastore.Has
func (d *Store) Has(ctx context.Context, key ds.Key) (exists bool, err error) {
	d.syncLock.Lock()
	defer d.syncLock.Unlock()

	_, found := d.values[key]
	return found, nil
}

func (d *Store) NewTransaction(ctx context.Context, readOnly bool) (ds.Txn, error) {
	return NewTransaction(d, readOnly), nil
}

// Put implements Datastore.Put
func (d *Store) Put(ctx context.Context, key ds.Key, value []byte) (err error) {
	d.syncLock.Lock()
	defer d.syncLock.Unlock()

	d.values[key] = value
	return nil
}

// Query implements Datastore.Query
func (d *Store) Query(ctx context.Context, q dsq.Query) (dsq.Results, error) {
	d.syncLock.Lock()
	defer d.syncLock.Unlock()

	re := make([]dsq.Entry, 0, len(d.values))
	for k, v := range d.values {
		e := dsq.Entry{Key: k.String(), Size: len(v)}
		if !q.KeysOnly {
			e.Value = v
		}
		re = append(re, e)
	}
	r := dsq.ResultsWithEntries(q, re)
	r = dsq.NaiveQueryApply(q, r)
	return r, nil
}

// Sync implements Datastore.Sync
func (d *Store) Sync(ctx context.Context, prefix ds.Key) error {
	return nil
}

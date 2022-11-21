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

// Datastore uses a btree for internal storage.
type Datastore struct {
	km     *keyMutex
	values *btree.Map[string, []byte]
}

var _ ds.Datastore = (*Datastore)(nil)
var _ ds.Batching = (*Datastore)(nil)
var _ ds.TxnFeature = (*Datastore)(nil)

// NewDatastore constructs an empty Datastore.
func NewDatastore() (d *Datastore) {
	return &Datastore{
		km:     newKeyMutex(),
		values: btree.NewMap[string, []byte](2),
	}
}

// Batch return a ds.Batch datastore based on Datastore
func (d *Datastore) Batch(ctx context.Context) (ds.Batch, error) {
	return newBasicBatch(d), nil
}

func (d *Datastore) Close() error {
	d.km.close <- struct{}{}
	return nil
}

// Delete implements ds.Delete
func (d *Datastore) Delete(ctx context.Context, key ds.Key) (err error) {
	d.km.lock(key.String())
	defer d.km.unlock(key.String())

	d.values.Delete(key.String())
	return nil
}

// Get implements ds.Get
func (d *Datastore) Get(ctx context.Context, key ds.Key) (value []byte, err error) {
	d.km.rlock(key.String())
	defer d.km.runlock(key.String())

	if val, exists := d.values.Get(key.String()); exists {
		return val, nil
	}
	return nil, ds.ErrNotFound
}

// GetSize implements ds.GetSize
func (d *Datastore) GetSize(ctx context.Context, key ds.Key) (size int, err error) {
	d.km.rlock(key.String())
	defer d.km.runlock(key.String())

	if val, exists := d.values.Get(key.String()); exists {
		return len(val), nil
	}
	return 0, ds.ErrNotFound
}

// Has implements ds.Has
func (d *Datastore) Has(ctx context.Context, key ds.Key) (exists bool, err error) {
	d.km.rlock(key.String())
	defer d.km.runlock(key.String())

	_, exists = d.values.Get(key.String())
	return exists, nil
}

// NewTransaction return a ds.Txn datastore based on Datastore
func (d *Datastore) NewTransaction(ctx context.Context, readOnly bool) (ds.Txn, error) {
	return newTransaction(d, readOnly), nil
}

// Put implements ds.Put
func (d *Datastore) Put(ctx context.Context, key ds.Key, value []byte) (err error) {
	d.km.lock(key.String())
	defer d.km.unlock(key.String())

	d.values.Set(key.String(), value)
	return nil
}

// Query implements ds.Query
func (d *Datastore) Query(ctx context.Context, q dsq.Query) (dsq.Results, error) {
	d.km.querymu.RLock()
	defer d.km.querymu.RUnlock()

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
func (d *Datastore) Sync(ctx context.Context, prefix ds.Key) error {
	return nil
}

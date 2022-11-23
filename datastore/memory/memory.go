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
	"time"

	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	"github.com/tidwall/btree"
)

type item struct {
	key       string
	version   uint64
	val       []byte
	isDeleted bool
}

func byKeys(a, b item) bool {
	switch {
	case a.key < b.key:
		return true
	case a.key == b.key && a.version < b.version:
		return true
	default:
		return false
	}
}

var dbV *uint64

func init() {
	v := uint64(0)
	dbV = &v
}

// Datastore uses a btree for internal storage.
type Datastore struct {
	version  *uint64
	values   *btree.BTreeG[item]
	name     uint64
	compress chan struct{}
	close    chan struct{}
}

var _ ds.Datastore = (*Datastore)(nil)
var _ ds.Batching = (*Datastore)(nil)
var _ ds.TxnFeature = (*Datastore)(nil)

// NewDatastore constructs an empty Datastore.
func NewDatastore(ctx context.Context) *Datastore {
	v := uint64(0)
	d := &Datastore{
		values:   btree.NewBTreeG(byKeys),
		version:  &v,
		name:     atomic.AddUint64(dbV, 1),
		compress: make(chan struct{}),
		close:    make(chan struct{}),
	}
	go d.compressor(ctx)
	return d
}

// Batch return a ds.Batch datastore based on Datastore
func (d *Datastore) Batch(ctx context.Context) (ds.Batch, error) {
	return newBasicBatch(d), nil
}

func (d *Datastore) Close() error {
	d.close <- struct{}{}
	return nil
}

// Delete implements ds.Delete
func (d *Datastore) Delete(ctx context.Context, key ds.Key) (err error) {
	v := atomic.AddUint64(d.version, 1)
	d.values.Set(item{key: key.String(), version: v, isDeleted: true})
	return nil
}

// Get implements ds.Get
func (d *Datastore) Get(ctx context.Context, key ds.Key) (value []byte, err error) {
	result := item{}
	d.values.Descend(item{key: key.String(), version: atomic.LoadUint64(d.version)}, func(item item) bool {
		if key.String() == item.key && !item.isDeleted {
			result = item
		}
		return false
	})
	if result.key == "" {
		return nil, ds.ErrNotFound
	}
	return result.val, nil
}

// GetSize implements ds.GetSize
func (d *Datastore) GetSize(ctx context.Context, key ds.Key) (size int, err error) {
	result := item{}
	d.values.Descend(item{key: key.String(), version: atomic.LoadUint64(d.version)}, func(item item) bool {
		if key.String() == item.key && !item.isDeleted {
			result = item
		}
		return false
	})
	if result.key == "" {
		return 0, ds.ErrNotFound
	}
	return len(result.val), nil
}

// Has implements ds.Has
func (d *Datastore) Has(ctx context.Context, key ds.Key) (exists bool, err error) {
	result := item{}
	d.values.Descend(item{key: key.String(), version: atomic.LoadUint64(d.version)}, func(item item) bool {
		if key.String() == item.key && !item.isDeleted {
			result = item
		}
		return false
	})
	return result.key != "", nil
}

// NewTransaction return a ds.Txn datastore based on Datastore
func (d *Datastore) NewTransaction(ctx context.Context, readOnly bool) (ds.Txn, error) {
	return newTransaction(d, readOnly), nil
}

// Put implements ds.Put
func (d *Datastore) Put(ctx context.Context, key ds.Key, value []byte) (err error) {
	v := atomic.AddUint64(d.version, 1)
	d.values.Set(item{key: key.String(), version: v, val: value})
	return nil
}

// Query implements ds.Query
func (d *Datastore) Query(ctx context.Context, q dsq.Query) (dsq.Results, error) {
	re := make([]dsq.Entry, 0, d.values.Len())
	iter := d.values.Iter()
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

		if item.isDeleted {
			continue
		}

		e := dsq.Entry{Key: item.key, Size: len(item.val)}
		if !q.KeysOnly {
			e.Value = item.val
		}

		re = append(re, e)
	}
	iter.Release()

	r := dsq.ResultsWithEntries(q, re)
	r = dsq.NaiveQueryApply(q, r)
	return r, nil
}

// Sync implements ds.Sync
func (d *Datastore) Sync(ctx context.Context, prefix ds.Key) error {
	return nil
}

func (d *Datastore) compressor(ctx context.Context) {
	now := time.Now()
	nextCompression := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	for {
		select {
		case <-d.close:
			return
		case <-d.compress:
			d.smash(ctx)
		case <-time.After(1 * time.Minute):
			now := time.Now()
			if now.After(nextCompression) {
				d.smash(ctx)
				nextCompression = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			}
		}
	}
}

func (d *Datastore) smash(ctx context.Context) {
	itemsToDelete := []item{}
	iter := d.values.Iter()
	iter.Next()
	item := iter.Item()
	// fast forward to last inserted version and delete versions before it
	for iter.Next() {
		if item.key == iter.Item().key {
			itemsToDelete = append(itemsToDelete, item)
		}
		item = iter.Item()
	}
	iter.Release()

	for _, i := range itemsToDelete {
		d.values.Delete(i)
	}
}

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
	"time"

	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	"github.com/tidwall/btree"
)

type dsTxn struct {
	dsVersion  uint64
	txnVersion uint64
	expiresAt  time.Time
	txn        *basicTxn
}

func byDSVersion(a, b dsTxn) bool {
	switch {
	case a.dsVersion < b.dsVersion:
		return true
	case a.dsVersion == b.dsVersion && a.txnVersion < b.txnVersion:
		return true
	default:
		return false
	}
}

type dsItem struct {
	key       string
	version   uint64
	val       []byte
	isDeleted bool
	isGet     bool
}

func byKeys(a, b dsItem) bool {
	switch {
	case a.key < b.key:
		return true
	case a.key == b.key && a.version < b.version:
		return true
	default:
		return false
	}
}

// Datastore uses a btree for internal storage.
type Datastore struct {
	// Latest committed version.
	version     *uint64
	values      *btree.BTreeG[dsItem]
	inFlightTxn *btree.BTreeG[dsTxn]

	closing  chan struct{}
	closed   bool
	closeLk  sync.RWMutex
	commitLk sync.Mutex
}

var _ ds.Datastore = (*Datastore)(nil)
var _ ds.Batching = (*Datastore)(nil)
var _ ds.TxnFeature = (*Datastore)(nil)

// NewDatastore constructs an empty Datastore.
func NewDatastore(ctx context.Context) *Datastore {
	v := uint64(0)
	d := &Datastore{
		version:     &v,
		values:      btree.NewBTreeG(byKeys),
		inFlightTxn: btree.NewBTreeG(byDSVersion),
		closing:     make(chan struct{}),
	}
	go d.purgeOldVersions(ctx)
	go d.handleContextDone(ctx)
	return d
}

func (d *Datastore) getVersion() uint64 {
	return atomic.LoadUint64(d.version)
}

func (d *Datastore) nextVersion() uint64 {
	return atomic.AddUint64(d.version, 1)
}

// Batch return a ds.Batch datastore based on Datastore.
func (d *Datastore) Batch(ctx context.Context) (ds.Batch, error) {
	return d.newBasicBatch(), nil
}

// newBasicBatch returns a ds.Batch datastore.
func (d *Datastore) newBasicBatch() ds.Batch {
	return &basicBatch{
		ops: make(map[ds.Key]op),
		ds:  d,
	}
}

func (d *Datastore) Close() error {
	d.closeLk.Lock()
	defer d.closeLk.Unlock()
	if d.closed {
		return ErrClosed
	}

	d.closed = true
	close(d.closing)

	iter := d.inFlightTxn.Iter()

	for iter.Next() {
		iter.Item().txn.close()
	}
	iter.Release()

	return nil
}

// Delete implements ds.Delete.
func (d *Datastore) Delete(ctx context.Context, key ds.Key) (err error) {
	d.closeLk.RLock()
	defer d.closeLk.RUnlock()
	if d.closed {
		return ErrClosed
	}
	tx := d.newTransaction(false)
	// An error can never happen at this stage so we explicitly ignore it
	_ = tx.Delete(ctx, key)
	return tx.Commit(ctx)
}

func (d *Datastore) get(ctx context.Context, key ds.Key, version uint64) dsItem {
	result := dsItem{}
	d.values.Descend(dsItem{key: key.String(), version: version}, func(item dsItem) bool {
		if key.String() == item.key {
			result = item
		}
		// We only care about the last version so we stop iterating right away by returning false.
		return false
	})
	return result
}

// Get implements ds.Get.
func (d *Datastore) Get(ctx context.Context, key ds.Key) (value []byte, err error) {
	d.closeLk.RLock()
	defer d.closeLk.RUnlock()
	if d.closed {
		return nil, ErrClosed
	}
	result := d.get(ctx, key, d.getVersion())
	if result.key == "" || result.isDeleted {
		return nil, ds.ErrNotFound
	}
	return result.val, nil
}

// GetSize implements ds.GetSize.
func (d *Datastore) GetSize(ctx context.Context, key ds.Key) (size int, err error) {
	d.closeLk.RLock()
	defer d.closeLk.RUnlock()
	if d.closed {
		return 0, ErrClosed
	}
	result := d.get(ctx, key, d.getVersion())
	if result.key == "" || result.isDeleted {
		return 0, ds.ErrNotFound
	}
	return len(result.val), nil
}

// Has implements ds.Has.
func (d *Datastore) Has(ctx context.Context, key ds.Key) (exists bool, err error) {
	d.closeLk.RLock()
	defer d.closeLk.RUnlock()
	if d.closed {
		return false, ErrClosed
	}
	result := d.get(ctx, key, d.getVersion())
	return result.key != "" && !result.isDeleted, nil
}

// NewTransaction return a ds.Txn datastore based on Datastore.
func (d *Datastore) NewTransaction(ctx context.Context, readOnly bool) (ds.Txn, error) {
	d.closeLk.RLock()
	defer d.closeLk.RUnlock()
	if d.closed {
		return nil, ErrClosed
	}
	return d.newTransaction(readOnly), nil
}

// newTransaction returns a ds.Txn datastore.
//
// isInternal should be set to true if this transaction is created from within the
// datastore and is already protected by stuff like locks.  Failure to correctly set
// this to true may result in deadlocks.  Failure to correctly set it to false may lead
// to other concurrency issues.
func (d *Datastore) newTransaction(readOnly bool) ds.Txn {
	v := d.getVersion()
	txn := &basicTxn{
		ops:       btree.NewBTreeG(byKeys),
		ds:        d,
		readOnly:  readOnly,
		dsVersion: &v,
	}

	d.inFlightTxn.Set(dsTxn{v, v + 1, time.Now().Add(1 * time.Hour), txn})
	return txn
}

// Put implements ds.Put.
func (d *Datastore) Put(ctx context.Context, key ds.Key, value []byte) (err error) {
	d.closeLk.RLock()
	defer d.closeLk.RUnlock()
	if d.closed {
		return ErrClosed
	}
	tx := d.newTransaction(false)
	// An error can never happen at this stage so we explicitly ignore it
	_ = tx.Put(ctx, key, value)
	return tx.Commit(ctx)
}

// Query implements ds.Query.
func (d *Datastore) Query(ctx context.Context, q dsq.Query) (dsq.Results, error) {
	d.closeLk.RLock()
	defer d.closeLk.RUnlock()
	if d.closed {
		return nil, ErrClosed
	}
	re := make([]dsq.Entry, 0, d.values.Height())
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

// Sync implements ds.Sync.
func (d *Datastore) Sync(ctx context.Context, prefix ds.Key) error {
	d.closeLk.RLock()
	defer d.closeLk.RUnlock()
	if d.closed {
		return ErrClosed
	}
	return nil
}

// purgeOldVersions will execute the purge once a day or when explicitly requested.
func (d *Datastore) purgeOldVersions(ctx context.Context) {
	dbStartTime := time.Now()
	nextCompression := time.Date(dbStartTime.Year(), dbStartTime.Month(), dbStartTime.Day()+1,
		0, 0, 0, 0, dbStartTime.Location())

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.closing:
			return
		case <-time.After(time.Until(nextCompression)):
			d.executePurge(ctx)
			now := time.Now()
			nextCompression = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		}
	}
}

func (d *Datastore) executePurge(ctx context.Context) {
	// purging bellow this version
	v := d.getVersion()
	if dsTxn, hasMin := d.inFlightTxn.Min(); hasMin {
		v = dsTxn.dsVersion
	}

	for {
		itemsToDelete := []dsItem{}
		iter := d.values.Iter()
		iter.Next()
		item := iter.Item()
		// fast forward to last inserted version and delete versions before it
		total := 0
		for iter.Next() {
			if iter.Item().version > v {
				continue
			}
			if item.key == iter.Item().key {
				itemsToDelete = append(itemsToDelete, item)
				total++
			}
			item = iter.Item()
			// we don't want to delete more than 1000 items at a time
			// to prevent loading too much into memory
			if total >= 1000 {
				break
			}
		}
		iter.Release()

		if total == 0 {
			return
		}

		for _, i := range itemsToDelete {
			d.values.Delete(i)
		}
	}
}

func (d *Datastore) handleContextDone(ctx context.Context) {
	select {
	case <-d.closing:
		return
	case <-ctx.Done():
		// It is safe to ignore the error since the only error that could occur is if the
		// datastore is already closed, in which case the purpose of the `Close` call is already covered.
		_ = d.Close()
	}
}

// commit commits the given transaction to the datastore.
//
// WARNING: This is a notable bottleneck, as commits can only be commited one at a time (handled internally).
// This is to ensure correct, threadsafe, mututation of the datastore version.
func (d *Datastore) commit(ctx context.Context, t *basicTxn) error {
	d.commitLk.Lock()
	defer d.commitLk.Unlock()

	// The commitLk scope must include checkForConflicts, and it must be a write lock. The datastore version
	// cannot be allowed to change between here and the release of the iterator, else the check for conflicts
	// will be stale and potentially out of date.
	err := t.checkForConflicts(ctx)
	if err != nil {
		return err
	}

	iter := t.ops.Iter()
	v := t.ds.nextVersion()
	for iter.Next() {
		if iter.Item().isGet {
			continue
		}
		item := iter.Item()
		item.version = v
		t.ds.values.Set(item)
	}
	iter.Release()
	return nil
}

func (d *Datastore) clearOldInFlightTxn(ctx context.Context) {
	if d.inFlightTxn.Height() == 0 {
		return
	}

	now := time.Now()
	for {
		itemsToDelete := []dsTxn{}
		iter := d.inFlightTxn.Iter()

		total := 0
		for iter.Next() {
			if now.After(iter.Item().expiresAt) {
				itemsToDelete = append(itemsToDelete, iter.Item())
				total++
			}
		}
		iter.Release()

		if total == 0 {
			return
		}

		for _, i := range itemsToDelete {
			d.inFlightTxn.Delete(i)
		}
	}
}

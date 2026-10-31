// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/db/description"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/db/vectorindex"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// deferredBatchSize is how many pending entries one drain transaction applies at first. Benchmarks
// showed no speed difference between 16 and 256.
const deferredBatchSize = 32

// vectorMaintenancePendingLimit is how many pending entries an index can hold before writes stop
// deferring and update the graph inline. This makes a fast writer slow down instead of letting
// the pending list grow forever. Every search scans all pending entries, and 1000 of them add
// only about 0.6 ms (128 dims) to 3.6 ms (1536 dims). A lower limit pushes write bursts back to
// slow inline writes much sooner. It is a var so tests can change it.
var vectorMaintenancePendingLimit int64 = 1000

// pendingLimit returns the pending count at which an index stops deferring.
func pendingLimit() int64 {
	return max(1, vectorMaintenancePendingLimit)
}

// deferredBatchesPerPass is how many batches the writer applies to one index before it moves on to
// the next dirty index. Without it, an index with steady writes refills every batch, so the writer
// never leaves it and the other indexes fill up and fall back to slow inline writes.
const deferredBatchesPerPass = 8

// vectorMaintenanceRetryDelay is how long the writer waits before it first retries a failed index.
// Without it the writer would spin on an error. It is a var so tests can shorten it.
var vectorMaintenanceRetryDelay = 200 * time.Millisecond

// vectorMaintenanceMaxRetryDelay caps the retry delay. The delay doubles on every failed pass, so
// an error that does not go away is retried (and logged) every half minute, not five times a second.
var vectorMaintenanceMaxRetryDelay = 30 * time.Second

// vectorMaintenanceGate is a test hook that runs before each drain batch. Tests use it to hold
// the writer, to change the index mid-drain, or to make a drain fail by returning an error. It is
// nil in production.
var vectorMaintenanceGate func(ctx context.Context, collectionID string, indexID uint32) error

// afterPendingRead is a test hook that runs right after a drain batch reads its entries. A test
// can use it to overwrite an entry before the batch commits. It is nil in production.
var afterPendingRead func()

// suppressVectorMaintenanceRun makes newDB build the writer without starting its run loop. Tests
// use it to keep pending entries in place or to drain by hand. Only tests should set it. It is safe
// because the db package runs its tests one at a time.
var suppressVectorMaintenanceRun bool

// vectorIndexRef names one vector index in the writer's in-memory maps. It is never stored.
type vectorIndexRef struct {
	collectionID string
	indexID      uint32
}

// vectorMaintenanceWriter applies the pending entries of deferred vector indexes to their graphs
// in the background. There is one per DB and it has a single goroutine, so two writers never touch
// the same index.
//
// Writes only wake it up. It finds the work by reading the stored pending entries, so a missed or
// merged wake loses nothing. After a restart every deferred index is marked dirty again.
type vectorMaintenanceWriter struct {
	db *DB

	// indexes holds the *deferredIndexState of every deferred index the writer knows about. Document
	// writes on any goroutine read and update it, so everything in it is atomic and no write takes
	// a lock.
	indexes sync.Map

	// wake has room for one signal, so many writes only cause one wake up.
	wake chan struct{}

	// retryDelay and maxRetryDelay are copied from the package vars when the writer is made, so the
	// run loop never reads those vars while a test changes them for another DB.
	retryDelay    time.Duration
	maxRetryDelay time.Duration
}

// deferredIndexState is what the writer keeps in memory about one deferred index.
type deferredIndexState struct {
	// pending is an approximate count of pending writes. It is used to slow writers down. It can be
	// too high when a write overwrites an entry that is still pending. A full drain resets it, and
	// after a restart it starts from the stored count.
	pending atomic.Int64
	// dirty is set when the index might have pending entries.
	dirty atomic.Bool
	// batchSize is the batch size when a smaller one was needed to fit Badger's transaction limit.
	// Zero means deferredBatchSize.
	batchSize atomic.Int64
}

func newVectorMaintenanceWriter(db *DB) *vectorMaintenanceWriter {
	return &vectorMaintenanceWriter{
		db:            db,
		wake:          make(chan struct{}, 1),
		retryDelay:    vectorMaintenanceRetryDelay,
		maxRetryDelay: vectorMaintenanceMaxRetryDelay,
	}
}

// state returns the in-memory state of an index, creating it if needed.
func (w *vectorMaintenanceWriter) state(key vectorIndexRef) *deferredIndexState {
	s, ok := w.indexes.Load(key)
	if !ok {
		s, _ = w.indexes.LoadOrStore(key, &deferredIndexState{})
	}
	// Only *deferredIndexState values are ever stored.
	state, _ := s.(*deferredIndexState)
	return state
}

// pendingCounter returns the pending counter for an index, creating it if needed.
func (w *vectorMaintenanceWriter) pendingCounter(key vectorIndexRef) *atomic.Int64 {
	return &w.state(key).pending
}

// overLimit reports whether the index has so many pending writes that new writes should update
// the graph inline instead.
func (w *vectorMaintenanceWriter) overLimit(key vectorIndexRef) bool {
	return w.pendingCounter(key).Load() >= pendingLimit()
}

// noteWrite is called after a deferred write commits. It counts the write and wakes the writer.
func (w *vectorMaintenanceWriter) noteWrite(key vectorIndexRef) {
	s := w.state(key)
	s.pending.Add(1)
	s.dirty.Store(true)
	w.notify()
}

// markDirty marks an index as needing a drain and wakes the writer. It never blocks.
func (w *vectorMaintenanceWriter) markDirty(key vectorIndexRef) {
	w.state(key).dirty.Store(true)
	w.notify()
}

func (w *vectorMaintenanceWriter) notify() {
	select {
	case w.wake <- struct{}{}:
	default:
		// A wake is already waiting, so one more would add nothing.
	}
}

// takeDirty returns the dirty indexes and clears their flags. If a write commits while we drain, it
// marks its index dirty again, so the next pass picks it up.
func (w *vectorMaintenanceWriter) takeDirty() []vectorIndexRef {
	var out []vectorIndexRef
	w.indexes.Range(func(k, v any) bool {
		key, _ := k.(vectorIndexRef)
		s, _ := v.(*deferredIndexState)
		if s.dirty.CompareAndSwap(true, false) {
			out = append(out, key)
		}
		return true
	})
	return out
}

// run drains the dirty indexes each time it wakes, until ctx is cancelled. It starts by marking
// every deferred index dirty. Pending entries are stored on disk, so this picks up work left over
// from before a crash or restart. If a drain fails, the index stays dirty and run tries again
// later, waiting twice as long after each failed pass.
func (w *vectorMaintenanceWriter) run(ctx context.Context) {
	// The timer starts stopped. Reset arms it when a retry is needed.
	retry := time.NewTimer(time.Hour)
	retry.Stop()
	defer retry.Stop()

	delay := w.retryDelay
	scanned := false
	for {
		// Without the startup scan, work left from before a restart would wait for the next write
		// to its index, so a failed scan is retried like a failed drain.
		if !scanned {
			err := w.markAllDeferredDirty(ctx)
			if ctx.Err() != nil {
				return
			}
			if err != nil {
				log.ErrorE("Failed to list deferred vector indexes", err)
			}
			scanned = err == nil
		}

		// drainPass logs its own failures, so here the error only decides whether to retry.
		failed := w.drainPass(ctx) != nil || !scanned
		if failed {
			retry.Reset(delay)
			delay = min(2*delay, w.maxRetryDelay)
		} else {
			delay = w.retryDelay
		}

		select {
		case <-ctx.Done():
			return
		case <-w.wake:
		case <-retry.C:
		}
	}
}

// markAllDeferredDirty marks every deferred vector index, in every collection version, as dirty.
func (w *vectorMaintenanceWriter) markAllDeferredDirty(ctx context.Context) error {
	rawTxn, err := w.db.NewTxn(true)
	if err != nil {
		return err
	}
	defer rawTxn.Discard()

	versions, err := description.GetCollections(InitContext(ctx, rawTxn), w.db.collectionRepository)
	if err != nil {
		return err
	}
	txnCtx := InitContext(ctx, rawTxn)
	// Collection versions share their indexes, so the same index shows up once per version.
	seen := map[vectorIndexRef]struct{}{}
	for _, def := range versions {
		for _, idx := range def.Indexes {
			vd, ok := idx.GetVector()
			if !ok || vd.Maintenance != client.VectorIndexMaintenanceDeferred {
				continue
			}
			key := vectorIndexRef{def.CollectionID, idx.ID}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			// The pending count lives in memory, so after a restart it starts from what is stored.
			// Otherwise every restart would let another full limit of entries pile up. Add rather
			// than set, so writes that committed after this snapshot keep their count. A write
			// counted twice only makes the count too high until the next full drain.
			n, err := countPending(txnCtx, key)
			if err != nil {
				return err
			}
			w.pendingCounter(key).Add(int64(n))
			w.markDirty(key)
		}
	}
	return nil
}

// pendingKeyspace is where an index's pending entries are stored.
type pendingKeyspace struct {
	shortID uint32
	epoch   uint32
}

// pendingKeyspaceOf returns where an index's pending entries are stored. found is false when the
// collection or index has been removed.
func pendingKeyspaceOf(ctx context.Context, key vectorIndexRef) (ks pendingKeyspace, found bool, err error) {
	ks.shortID, err = id.GetCollectionShortID(ctx, key.collectionID)
	if errors.Is(err, corekv.ErrNotFound) {
		return ks, false, nil
	}
	if err != nil {
		return ks, false, err
	}
	ks.epoch, err = getIndexEpoch(ctx, key.collectionID, key.indexID)
	if errors.Is(err, fetcher.ErrIndexEpochNotFound) {
		return ks, false, nil
	}
	if err != nil {
		return ks, false, err
	}
	return ks, true, nil
}

// listPending returns up to limit pending entries of an index (all of them when limit is 0), and
// where they are stored. found is false when the collection or index has been removed, which
// callers treat as nothing to do rather than an error.
func listPending(
	ctx context.Context,
	key vectorIndexRef,
	limit int,
) (ks pendingKeyspace, entries []vectorindex.PendingEntry, found bool, err error) {
	ks, found, err = pendingKeyspaceOf(ctx, key)
	if err != nil || !found {
		return ks, nil, found, err
	}
	entries, err = vectorindex.ListPending(ctx, ks.shortID, key.indexID, ks.epoch, limit)
	if err != nil {
		return ks, nil, false, err
	}
	return ks, entries, true, nil
}

// countPending returns how many pending entries an index has, reading keys only. A removed
// collection or index has none.
func countPending(ctx context.Context, key vectorIndexRef) (int, error) {
	ks, found, err := pendingKeyspaceOf(ctx, key)
	if err != nil || !found {
		return 0, err
	}
	return vectorindex.CountPending(ctx, ks.shortID, key.indexID, ks.epoch)
}

// drainPass drains each dirty index once and returns the first error. Indexes that failed stay
// dirty. Only the run loop calls it, so two passes never overlap.
func (w *vectorMaintenanceWriter) drainPass(ctx context.Context) error {
	var firstErr error
	for _, key := range w.takeDirty() {
		// The context is only cancelled at shutdown. The dirty set is lost then anyway, and a restart
		// marks every deferred index dirty again, so there is nothing to keep.
		if err := ctx.Err(); err != nil {
			return err
		}
		err := w.drainIndex(ctx, key)
		if err == nil {
			continue
		}
		// A drain cut short by shutdown is normal, so do not report it.
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// Keep the index dirty so it gets retried.
		w.state(key).dirty.Store(true)
		if firstErr == nil {
			firstErr = err
		}
		// Inline writes (used when too many entries are pending) can clash with us. That is
		// normal under load and the retry fixes it, so we do not log it.
		if errors.Is(err, corekv.ErrTxnConflict) {
			continue
		}
		log.ErrorE("Failed to apply pending vector index writes", err,
			corelog.String("collectionID", key.collectionID),
			corelog.Any("indexID", key.indexID),
		)
	}
	return firstErr
}

// batchLimit returns the current batch size of an index.
func (w *vectorMaintenanceWriter) batchLimit(key vectorIndexRef) int {
	if n := w.state(key).batchSize.Load(); n > 0 {
		return int(n)
	}
	return deferredBatchSize
}

// setBatchLimit remembers a smaller batch size for an index.
func (w *vectorMaintenanceWriter) setBatchLimit(key vectorIndexRef, n int) {
	w.state(key).batchSize.Store(int64(n))
}

// drainIndex applies the index's pending entries in batches until none are left, or until it has
// applied deferredBatchesPerPass full batches. In that case it marks the index dirty again and
// returns, so the writer serves the other dirty indexes before it comes back to this one. If the
// index is gone or should not be maintained, it is skipped with no error.
func (w *vectorMaintenanceWriter) drainIndex(ctx context.Context, key vectorIndexRef) error {
	_, desc, err := w.db.findIndexDefinition(ctx, keys.IndexStateKey{
		CollectionID: key.collectionID,
		IndexID:      key.indexID,
	})
	if err != nil {
		if errors.Is(err, ErrIndexWithIDDoesNotExist) {
			w.forget(key)
			return nil
		}
		return err
	}
	vd, ok := desc.GetVector()
	// Can't happen today: only deferred vector indexes get marked dirty, indexes can't be altered,
	// and index IDs are never reused. This guards against a future way to change the mode in place.
	if !ok || vd.Maintenance != client.VectorIndexMaintenanceDeferred {
		w.forget(key)
		return nil
	}

	limit := w.batchLimit(key)
	counter := w.pendingCounter(key)
	for batches := 0; ; batches++ {
		if batches == deferredBatchesPerPass {
			// markDirty also wakes the run loop, so it starts the next pass right away.
			w.markDirty(key)
			return nil
		}
		if gate := vectorMaintenanceGate; gate != nil {
			if err := gate(ctx, key.collectionID, key.indexID); err != nil {
				return err
			}
		}
		if err := ctx.Err(); err != nil {
			return err
		}

		// Read the counter before scanning. Writes that land after this point may not be in the
		// scan, so we must not subtract them.
		seen := counter.Load()
		n, forget, err := w.drainBatch(ctx, key, vd, limit)
		if err != nil {
			// We react to the store's size error instead of estimating the size of a batch. This
			// keeps the writer independent of both the index algorithm and the storage engine.
			if errors.Is(err, corekv.ErrTxnTooBig) && limit > 1 {
				limit /= 2
				w.setBatchLimit(key, limit)
				continue
			}
			// Inline writes (used when the pending list is full) touch the same graph nodes as our
			// batch. A smaller batch commits faster and touches fewer nodes, so it is less likely to
			// clash again. Unlike the size limit this depends on load, so it is not remembered.
			if errors.Is(err, corekv.ErrTxnConflict) && limit > 1 {
				limit /= 2
				continue
			}
			return err
		}
		if forget {
			w.forget(key)
			return nil
		}
		if n < limit {
			// The list is empty. Subtract what the scan saw instead of resetting to zero, so we
			// keep the count of writes that came after. Those writes also marked the index dirty.
			subPending(counter, seen)
			return nil
		}
		subPending(counter, int64(n))
	}
}

// forget drops the state of an index the writer no longer maintains. Without this, every dropped
// index would leave it in memory until the node restarts.
func (w *vectorMaintenanceWriter) forget(key vectorIndexRef) {
	w.indexes.Delete(key)
}

// subPending subtracts n from the counter and never goes below zero.
func subPending(counter *atomic.Int64, n int64) {
	for {
		cur := counter.Load()
		if counter.CompareAndSwap(cur, max(cur-n, 0)) {
			return
		}
	}
}

// drainBatch applies and deletes up to limit pending entries in one transaction. It returns how
// many it handled. forget is true when the index should be skipped because it is dropped, failed
// or still building.
//
// The transaction reads every entry it deletes. If a document write overwrites one of them after
// that read, the commit conflicts and withTxnRetries runs the batch again with the newer entry.
// This way we never delete an entry we did not apply.
func (w *vectorMaintenanceWriter) drainBatch(
	ctx context.Context,
	key vectorIndexRef,
	vd *client.VectorIndexDescription,
	limit int,
) (n int, forget bool, err error) {
	err = w.db.withTxnRetries(ctx, func(txnCtx context.Context) error {
		n, forget = 0, false

		state, err := getIndexState(txnCtx, key.collectionID, key.indexID)
		switch {
		case err == nil:
			// While the index is building, writes go straight into the graph and the build picks up
			// every existing document, so there is nothing pending for us. Failed or dropping indexes
			// are not maintained.
			if state.isBuilding() || state.isFailed() || state.isDropping() {
				forget = true
				return nil
			}
		case !errors.Is(err, corekv.ErrNotFound):
			return err
		}

		ks, entries, found, err := listPending(txnCtx, key, limit)
		if err != nil {
			return err
		}
		if !found {
			forget = true
			return nil
		}
		if hook := afterPendingRead; hook != nil {
			hook()
		}
		if len(entries) == 0 {
			return nil
		}

		idx, err := vectorindex.Open(txnCtx, ks.shortID, key.indexID, ks.epoch, *vd)
		if err != nil {
			return err
		}
		for _, e := range entries {
			// Always delete first. Deleting a node the graph does not have does nothing, and for
			// an upsert it removes the old vector so the insert puts in the new one.
			if err := idx.Delete(e.NodeID); err != nil {
				return err
			}
			if e.Op == vectorindex.PendingUpsert {
				if err := idx.Insert(e.NodeID, e.Vector); err != nil {
					return err
				}
			}
			if err := vectorindex.DeletePending(txnCtx, ks.shortID, key.indexID, ks.epoch, e.NodeID); err != nil {
				return err
			}
		}
		n = len(entries)
		return nil
	})
	return n, forget, err
}

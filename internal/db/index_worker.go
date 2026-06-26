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
	"time"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// indexBuildConcurrency bounds concurrent builds and drops across distinct indexes. A var so tests
// can adjust it.
var indexBuildConcurrency = 4

// indexBuildRetryDelay is the backoff before re-draining a build or drop interrupted by a
// transaction conflict, giving the racing writer time to commit first. A var so tests can shrink it.
var indexBuildRetryDelay = 100 * time.Millisecond

// indexBuildWorker drains pending index build and drop state records in the background, reusing the
// batched backfill and GC. It is shared by startup recovery and on-demand async NewIndex/DeleteIndex:
// both stage a record and commit, the commit publishes an action event, and the worker drains it.
//
// The event is only a wake-up hint; the worker re-derives its work from the persisted records. This
// is correct when the triggering transaction rolls back (no record), when several events arrive
// (coalesced into one drain), and on restart (startup drain == on-demand drain).
type indexBuildWorker struct {
	db *DB

	// sub receives ActionExecution events, the primary wake source.
	sub event.Subscription

	// wake is a size-one coalesced wake-up channel for the initial drain and non-bus callers.
	wake chan struct{}

	// inFlight holds the keys of indexes whose build or drop is running, so a drain does not start
	// a second one for the same index. Keyed by inFlightKey.
	inFlight sync.Map

	// sem bounds concurrent builds across indexes.
	sem chan struct{}

	// drainMu serialises drain passes so only one pass increments builds at a time, letting
	// drainSync await its own pass without racing a concurrent drain.
	drainMu sync.Mutex

	// builds counts dispatched build goroutines so a drain can be awaited. Incremented under drainMu.
	builds sync.WaitGroup
}

// newIndexBuildWorker constructs a worker subscribed to action events. The subscription is closed
// by db.events.Close.
func (db *DB) newIndexBuildWorker() (*indexBuildWorker, error) {
	sub, err := db.events.Subscribe(event.ActionExecutionName)
	if err != nil {
		return nil, err
	}
	return &indexBuildWorker{
		db:   db,
		sub:  sub,
		wake: make(chan struct{}, 1),
		sem:  make(chan struct{}, indexBuildConcurrency),
	}, nil
}

// inFlightKey identifies an in-flight build or drop by its index. Build and drop of the same index
// share one key so they are mutually exclusive: a whole-index drop range-deletes every epoch while
// assuming no writer is touching them, so it must not run while a backfill of the same index is
// still writing entries (which would leave orphaned entries past the drop). A rebuild's stale-epoch
// sweep is bounded strictly below the live epoch and never dispatched here, so it is unaffected.
func inFlightKey(key keys.IndexStateKey) string {
	return key.CollectionID + "/" + indexSubject(key.IndexID)
}

// run drains pending records on start and on each wake-up until ctx is cancelled. It never blocks
// on a build: builds run in the bounded pool while the loop keeps reading the subscription, so a
// busy build cannot stall the bus (a blocked subscriber blocks every publisher).
func (w *indexBuildWorker) run(ctx context.Context) {
	w.drain(ctx)

	for {
		select {
		case <-ctx.Done():
			// Let dispatched builds observe cancellation and return before exiting, so the worker
			// does not outlive the storage it writes to.
			w.builds.Wait()
			return

		case msg, ok := <-w.sub.Message():
			if !ok {
				w.builds.Wait()
				return
			}
			if w.isWakeEvent(msg) {
				w.drain(ctx)
			}

		case <-w.wake:
			w.drain(ctx)
		}
	}
}

// isWakeEvent reports whether an action event is a reason to drain: an index build or drop entering
// the in-progress state. Terminal statuses and non-index actions are ignored.
func (w *indexBuildWorker) isWakeEvent(msg event.Message) bool {
	exec, ok := msg.Data.(event.ActionExecution)
	if !ok {
		return false
	}
	return isIndexAction(exec.Action) && exec.Status == client.InProgressActionStatus
}

// notify is a non-blocking wake-up backstop for callers that do not go through the bus. The action
// event published on commit is the primary trigger.
func (w *indexBuildWorker) notify() {
	select {
	case w.wake <- struct{}{}:
	default:
		// A wake is already pending.
	}
}

// drain lists every index state record and dispatches each pending build or drop through the
// in-flight guard and bounded pool, then sweeps stale epochs. It returns once work is dispatched,
// not when builds finish. Safe to call repeatedly: records are the source of truth, the backfill
// resumes from its watermark, and the GC is idempotent.
func (w *indexBuildWorker) drain(ctx context.Context) {
	w.drainMu.Lock()
	defer w.drainMu.Unlock()
	w.drainLocked(ctx)
}

// drainLocked is the drain body. The caller must hold drainMu so only one pass increments builds.
func (w *indexBuildWorker) drainLocked(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}

	states, err := w.db.listAllIndexStates(ctx)
	if err != nil {
		log.ErrorE("Failed to list index states during drain", err)
		return
	}

	for _, rec := range states {
		if ctx.Err() != nil {
			return
		}
		switch {
		case rec.State.isBuilding():
			w.dispatch(ctx, rec.Key, client.BackfillIndexAction, func(c context.Context) error {
				return w.db.recoverBuilding(c, rec.Key, rec.State)
			})
		case rec.State.isDropping():
			w.dispatch(ctx, rec.Key, client.DropIndexAction, func(c context.Context) error {
				return w.db.recoverDropping(c, rec.Key)
			})
		default:
			// A failed or ready index requires no action.
		}
	}

	// A rebuild leaves superseded epochs with no record, and may crash after its build finishes
	// but before collecting them. Sweep every index; it is a no-op for an index with only its live
	// epoch, and the delete range is bounded strictly below the live epoch so it never touches an
	// in-flight build.
	if err := w.db.recoverStaleEpochs(ctx); err != nil {
		log.ErrorE("Failed to collect stale index epochs during drain", err)
	}
}

// drainSync runs a drain and blocks until every dispatched build or drop has finished. It is
// test-only: production never blocks on a build (see run). drainMu is held across both the dispatch
// and the wait so no concurrent drain increments builds meanwhile. The in-flight guard is shared
// with the background loop, so an index a prior pass is already building is not started twice;
// drainSync waits for that work too, since builds is shared.
func (w *indexBuildWorker) drainSync(ctx context.Context) {
	w.drainMu.Lock()
	defer w.drainMu.Unlock()
	w.drainLocked(ctx)
	w.builds.Wait()
}

// dispatch runs work for one index under the in-flight guard and the bounded pool. It skips the
// index if a build/drop is already in flight, otherwise it runs work in a goroutine that acquires a
// semaphore slot first. Acquiring the slot inside the goroutine keeps drain non-blocking when the
// pool is full.
func (w *indexBuildWorker) dispatch(
	ctx context.Context,
	key keys.IndexStateKey,
	action client.Action,
	work func(context.Context) error,
) {
	guardKey := inFlightKey(key)
	if _, loaded := w.inFlight.LoadOrStore(guardKey, struct{}{}); loaded {
		return // a build or drop for this index is already in flight
	}

	w.builds.Go(func() {
		// Re-drain after releasing the guard: another record for this index may have been
		// guard-skipped while this work ran (e.g. a drop staged while a build held the index), and
		// the build finishing is not otherwise a wake reason. The drain is a cheap no-op when
		// nothing is pending.
		defer w.notify()
		defer w.inFlight.Delete(guardKey)

		select {
		case w.sem <- struct{}{}:
		case <-ctx.Done():
			return
		}
		defer func() { <-w.sem }()

		if ctx.Err() != nil {
			return
		}
		err := work(ctx)
		if err == nil {
			return
		}
		log.ErrorE("Index build worker task failed", err,
			corelog.String("collectionID", key.CollectionID),
			corelog.Any("indexID", key.IndexID),
			corelog.Any("action", action),
		)
		// A transaction conflict is transient: the record is still in place and resumable from its
		// watermark, but nothing else re-drives it, so re-drain after a backoff. Without this the
		// index would stay building/dropping forever. A terminal error already recorded the failed
		// state and needs no retry.
		if errors.Is(err, corekv.ErrTxnConflict) {
			w.scheduleRetry(ctx)
		}
	})
}

// scheduleRetry wakes the worker after indexBuildRetryDelay so an index left building/dropping by a
// transaction conflict is re-drained and resumed. It is tracked under builds so shutdown waits for
// it, and returns early on ctx cancellation.
func (w *indexBuildWorker) scheduleRetry(ctx context.Context) {
	w.builds.Go(func() {
		timer := time.NewTimer(indexBuildRetryDelay)
		defer timer.Stop()
		select {
		case <-timer.C:
			w.notify()
		case <-ctx.Done():
		}
	})
}

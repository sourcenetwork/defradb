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
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// indexBuildConcurrency is the number of worker goroutines, so also the max builds/drops running at
// once. A var so tests can adjust it.
var indexBuildConcurrency = 4

// indexTaskQueueSize caps the pending-task queue. If it fills, a drain drops the task and re-derives
// it later from the records, so nothing is lost. Sized well above any real count of pending indexes.
var indexTaskQueueSize = 1024

// indexBuildRetryDelay is how long to wait before retrying a build or drop that hit a transaction
// conflict, giving the other writer time to commit first. A var so tests can shrink it.
var indexBuildRetryDelay = 100 * time.Millisecond

// indexBuildWorker runs pending index builds and drops in the background, reusing the batched
// backfill and GC. It serves both startup recovery and async NewIndex/DeleteIndex: each stages a
// record and commits, the commit fires an action event, and the worker picks it up.
//
// The event is just a hint to wake up; the worker figures out its work from the persisted records.
// That keeps it correct when the triggering txn rolls back (no record), when many events arrive at
// once (one drain handles all), and on restart (startup drain is the same as any drain).
type indexBuildWorker struct {
	db *DB

	// sub receives ActionExecution events, the main wake source.
	sub event.Subscription

	// wake is a size-one wake channel for the initial drain and callers not on the bus.
	wake chan struct{}

	// inFlight holds the indexes with a build or drop queued or running, so a drain won't queue a
	// second one for the same index. Keyed by inFlightKey.
	inFlight sync.Map

	// tasks is the pending build/drop queue. drain fills it; up to indexBuildConcurrency workers
	// drain it. This caps live goroutines at the pool size, not the number of pending indexes.
	tasks chan buildTask

	// liveWorkers counts running workers, so spawnWorker never starts more than indexBuildConcurrency.
	liveWorkers atomic.Int32

	// spawnMu makes spawnWorker's check-and-start atomic with a worker's exit check, so two drains
	// can't both spawn past the cap and a worker can't exit just as an enqueue counts it present.
	spawnMu sync.Mutex

	// drainMu lets only one drain enqueue at a time, so drainSync can wait for its own pass without a
	// concurrent drain racing it.
	drainMu sync.Mutex

	// builds counts running workers so a drain can wait for them (drainSync, Close).
	builds sync.WaitGroup

	// sweepRunning is set while a stale-epoch sweep goroutine is running, so a drain won't start a
	// second one. The running sweep re-scans every marker, so a skipped drain loses no work.
	sweepRunning atomic.Bool
}

// buildTask is one unit of work for a worker: the target index, the action (for logging), and the
// closure that does it.
type buildTask struct {
	key    keys.IndexStateKey
	action client.Action
	work   func(context.Context) error
}

// newIndexBuildWorker builds a worker subscribed to action events. db.events.Close closes the sub.
func (db *DB) newIndexBuildWorker() (*indexBuildWorker, error) {
	sub, err := db.events.Subscribe(event.ActionExecutionName)
	if err != nil {
		return nil, err
	}
	return &indexBuildWorker{
		db:    db,
		sub:   sub,
		wake:  make(chan struct{}, 1),
		tasks: make(chan buildTask, indexTaskQueueSize),
	}, nil
}

// inFlightKey keys an in-flight build or drop by index only, so build and drop of the same index
// share a key and can't run together. A whole-index drop range-deletes every epoch assuming no
// writer is active, so it must not overlap a backfill still writing entries (that would leave
// entries behind the dropped index). The rebuild stale-epoch sweep is bounded below the live epoch
// and never dispatched here, so it's unaffected.
func inFlightKey(key keys.IndexStateKey) string {
	return key.CollectionID + "/" + indexSubject(key.IndexID)
}

// run drains on start and on every wake until ctx is cancelled. drain only enqueues and returns, so
// the loop keeps reading the bus and stays responsive while the pool does the builds.
func (w *indexBuildWorker) run(ctx context.Context) {
	w.drain(ctx)

	for {
		select {
		case <-ctx.Done():
			// Wait for in-flight builds to observe cancellation and stop, so the worker doesn't
			// outlive the storage it writes to.
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

// isWakeEvent reports whether an event is worth a drain: an index build or drop going in-progress.
// Terminal statuses and non-index actions are ignored.
func (w *indexBuildWorker) isWakeEvent(msg event.Message) bool {
	exec, ok := msg.Data.(event.ActionExecution)
	if !ok {
		return false
	}
	return isIndexAction(exec.Action) && exec.Status == client.InProgressActionStatus
}

// notify is a non-blocking wake for callers not on the bus. The commit event is the main trigger.
func (w *indexBuildWorker) notify() {
	select {
	case w.wake <- struct{}{}:
	default:
		// A wake is already pending.
	}
}

// drain lists every index state record, queues each pending build or drop, then sweeps stale epochs.
// It returns once work is queued, not when it finishes. Safe to repeat: records are the source of
// truth, the backfill resumes from its watermark, and the GC is idempotent.
func (w *indexBuildWorker) drain(ctx context.Context) {
	w.drainMu.Lock()
	defer w.drainMu.Unlock()
	w.drainLocked(ctx)
}

// drainLocked is the drain body. The caller must hold drainMu.
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

	// A rebuild leaves superseded epochs with no record and may crash before collecting them. Sweep
	// them off the run loop: the GC can loop over many batches, and running it here would stop the
	// loop reading the event bus and stall publishers (the same reason builds go through the pool).
	w.dispatchSweep(ctx)
}

// dispatchSweep runs the stale-epoch sweep in a builds-tracked goroutine, unless one is already
// running. The sweep is a single global pass over the markers, not per-index, so one at a time is
// enough; the running pass re-scans every marker, and it re-drives on finish to catch markers that
// appeared while it ran.
func (w *indexBuildWorker) dispatchSweep(ctx context.Context) {
	if !w.sweepRunning.CompareAndSwap(false, true) {
		return
	}
	w.builds.Go(func() {
		defer w.sweepRunning.Store(false)
		swept, err := w.db.recoverStaleEpochs(ctx)
		if swept > 0 {
			// A marker committed while the sweep ran won't wake it again, so re-drive to catch it. Only
			// when this pass did work, or an idle drain would spin sweep -> notify -> drain forever.
			w.notify()
		}
		if err != nil {
			log.ErrorE("Failed to collect stale index epochs", err)
			// Stale epochs have no record to re-dispatch them, so on a conflict retry it ourselves;
			// otherwise they'd wait for an unrelated wake.
			if errors.Is(err, corekv.ErrTxnConflict) {
				w.scheduleRetry(ctx)
			}
		}
	})
}

// dispatch queues one index's work behind the in-flight guard, then makes sure a worker is running
// to pick it up. It skips the index if a build or drop is already queued or running. The enqueue is
// non-blocking, so drain never blocks; on a full queue the task is dropped and re-derived later.
func (w *indexBuildWorker) dispatch(
	ctx context.Context,
	key keys.IndexStateKey,
	action client.Action,
	work func(context.Context) error,
) {
	guardKey := inFlightKey(key)
	if _, loaded := w.inFlight.LoadOrStore(guardKey, struct{}{}); loaded {
		return // already queued or running for this index
	}

	select {
	case w.tasks <- buildTask{key: key, action: action, work: work}:
		w.spawnWorker(ctx)
	default:
		// Queue full: drop the task and release the guard; a later drain re-queues this index. Don't
		// notify. A full queue means every worker is busy, and each re-drives on completion (runTask's
		// deferred notify), so this index gets picked up when a slot frees. Notifying instead would
		// spin drain -> full -> notify at full CPU until then. The guard means one task per index, so
		// with the default size this only happens past thousands of pending indexes.
		w.inFlight.Delete(guardKey)
		log.InfoContext(ctx, "Index build queue full, deferring task to a later drain",
			corelog.String("collectionID", key.CollectionID),
			corelog.Any("indexID", key.IndexID),
			corelog.Any("action", action),
		)
	}
}

// spawnWorker starts a worker for the task just queued, unless the pool is already at the cap. It
// spawns one per enqueue rather than counting idle workers, because a busy worker looks the same as
// an idle one by count, and that would stop independent indexes from building in parallel. An extra
// worker that finds nothing to do exits right away, so the pool never goes over the cap. The
// check-and-start runs under spawnMu so two enqueues can't both spawn past it.
func (w *indexBuildWorker) spawnWorker(ctx context.Context) {
	w.spawnMu.Lock()
	defer w.spawnMu.Unlock()

	if w.liveWorkers.Load() >= int32(indexBuildConcurrency) {
		return
	}
	w.liveWorkers.Add(1)
	w.builds.Go(func() {
		w.consume(ctx)
	})
}

// consume runs queued tasks until the queue is empty or ctx is cancelled, then returns so the worker
// count drops back to zero when idle. The next drain refills the queue and re-spawns.
//
// The exit check runs under spawnMu and decrements liveWorkers there, making it atomic with
// spawnWorker's cap check: a concurrent enqueue either sees this worker still counted (and it takes
// the task on the re-check before exiting) or sees the decrement (and spawns a fresh worker). Either
// way a queued task always has a worker.
func (w *indexBuildWorker) consume(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			w.exitNow()
			return
		}
		select {
		case task := <-w.tasks:
			w.runTask(ctx, task)
		default:
			if w.exitIfEmpty() {
				return
			}
		}
	}
}

// exitIfEmpty decrements the worker count and returns true (the caller must return) if the queue is
// empty. If a task arrived after consume's non-blocking read, it returns false and the worker keeps
// going. Runs under spawnMu with spawnWorker.
func (w *indexBuildWorker) exitIfEmpty() bool {
	w.spawnMu.Lock()
	defer w.spawnMu.Unlock()
	if len(w.tasks) > 0 {
		return false
	}
	w.liveWorkers.Add(-1)
	return true
}

// exitNow decrements the worker count for a worker leaving on ctx cancellation, dropping any queued
// tasks (the next startup drain re-derives them from the records).
func (w *indexBuildWorker) exitNow() {
	w.spawnMu.Lock()
	defer w.spawnMu.Unlock()
	w.liveWorkers.Add(-1)
}

// runTask runs one build or drop, releases its guard, and re-drives. A conflict reschedules the
// index; a terminal error has already recorded the failed state.
func (w *indexBuildWorker) runTask(ctx context.Context, task buildTask) {
	// Re-drive after releasing the guard: a record for this index may have been skipped while this
	// ran (e.g. a drop staged behind the build), and finishing isn't otherwise a wake. Cheap no-op
	// when nothing is pending.
	defer w.notify()
	defer w.inFlight.Delete(inFlightKey(task.key))

	if ctx.Err() != nil {
		return
	}

	err := task.work(ctx)

	// Shutdown cancelled the build. The record stays and the next startup drain resumes it, so this is
	// expected. Return quietly with no retry, since the worker is stopping.
	if ctx.Err() != nil {
		return
	}

	// A conflict is temporary: the record is still there and resumes from its watermark, but nothing
	// else wakes it, so retry after a backoff or it stays building forever. A real error already
	// recorded the failed state and needs no retry. Log them apart so a retry isn't read as a failure.
	if errors.Is(err, corekv.ErrTxnConflict) {
		log.InfoContext(ctx, "Index build worker task hit a conflict, retrying",
			corelog.String("collectionID", task.key.CollectionID),
			corelog.Any("indexID", task.key.IndexID),
			corelog.Any("action", task.action),
		)
		w.scheduleRetry(ctx)
		return
	}
	if err != nil {
		log.ErrorE("Index build worker task failed", err,
			corelog.String("collectionID", task.key.CollectionID),
			corelog.Any("indexID", task.key.IndexID),
			corelog.Any("action", task.action),
		)
		return
	}

	// The async error contract returns nothing to the caller, so this transition to ready/dropped is
	// only visible here.
	log.InfoContext(ctx, "Index build worker task completed",
		corelog.String("collectionID", task.key.CollectionID),
		corelog.Any("indexID", task.key.IndexID),
		corelog.Any("action", task.action),
	)
}

// scheduleRetry wakes the worker after indexBuildRetryDelay to retry an index left building or
// dropping by a conflict. Tracked under builds so shutdown waits for it; returns early on cancel.
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

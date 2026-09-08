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

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

// indexBackfillBatchSize is the number of documents indexed per batch transaction during
// backfill. It is a fixed doc count sized for the worst case: an index key embeds its field
// value, which can approach the storage engine's per-key limit, so 100 entries stay well
// under the transaction size limit even when every value is near-maximal. For typical small
// fields this is conservative; a byte-budget batch (sized by accumulated entry bytes) would
// pack far more docs per transaction and is tracked as a follow-up. Tests lower it to exercise
// multi-batch runs.
var indexBackfillBatchSize = 100

// IndexBuildGate is a test hook called at each backfill batch boundary. Tests set it to block a
// build at the building state so that window is observable. It is nil in production (only the nil
// check runs) and exported so tests in other packages can install it.
var IndexBuildGate func(ctx context.Context, collectionID string, indexID uint32)

// withTxnRetries runs attempt with a fresh read-write transaction set on the context
// and commits it afterwards. When the attempt or the commit fails with a transaction
// conflict, the attempt is re-run with a new transaction, up to db.MaxTxnRetries() times;
// the last conflict error is returned if retries run out. Other errors abort immediately.
// At least one attempt is always made, even if MaxTxnRetries() returns 0.
func (db *DB) withTxnRetries(ctx context.Context, attempt func(ctx context.Context) error) error {
	var lastErr error
	max := db.MaxTxnRetries()
	if max < 1 {
		max = 1
	}
	for i := 0; i < max; i++ {
		rawTxn, err := db.NewTxn(false)
		if err != nil {
			return err
		}
		txn, ok := rawTxn.(*Txn)
		if !ok {
			return ErrUnexpectedTxnType
		}
		txnCtx := InitContext(ctx, txn)

		if err := attempt(txnCtx); err != nil {
			txn.Discard()
			if errors.Is(err, corekv.ErrTxnConflict) {
				lastErr = err
				continue
			}
			return err
		}

		if err := txn.Commit(); err != nil {
			txn.Discard()
			if errors.Is(err, corekv.ErrTxnConflict) {
				lastErr = err
				continue
			}
			return err
		}
		return nil
	}
	return lastErr
}

// isBuildInterrupted reports whether the build stopped for a reason that lets it resume later, rather
// than failing: a transaction conflict or ctx cancellation on shutdown. The record stays in place and
// the next drain resumes it, so it must not be marked failed.
func isBuildInterrupted(err error) bool {
	return errors.Is(err, corekv.ErrTxnConflict) ||
		errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded)
}

// backfillIndex builds an index by indexing every existing document in batched transactions, then
// deleting the build record to mark it ready. It is used both for a fresh index and for a rebuild
// filling a new epoch; in both cases the epoch is resolved from the index's sequence.
//
// startAfter resumes after the given document short ID from a persisted watermark; pass None to build the
// whole collection. A non-retryable error marks the index failed; a conflict leaves it resumable.
func (db *DB) backfillIndex(
	ctx context.Context,
	def client.CollectionVersion,
	desc client.IndexDescription,
	startAfter immutable.Option[uint64],
) error {
	builtEpoch, err := db.fillIndexBatches(ctx, def, desc, startAfter)
	if err != nil {
		return err
	}

	err = db.withTxnRetries(ctx, func(c context.Context) error {
		// A version switch can advance the epoch while this build runs; the build stays pinned to the
		// epoch it started on, which is now stale. Completing would delete the record the version switch
		// staged for the new epoch, leaving it unbuilt. Only complete if this build still filled the
		// live epoch; otherwise leave the record for the worker to build the new epoch from the top.
		liveEpoch, err := getIndexEpoch(c, def.CollectionID, desc.ID)
		if err != nil {
			return err
		}
		if liveEpoch != builtEpoch {
			return nil
		}
		return db.completeIndexBuild(c, def.CollectionID, desc.ID)
	})
	if err == nil {
		return nil
	}
	if isBuildInterrupted(err) {
		return NewErrIndexBackfillInterrupted(err, desc.Name)
	}
	markErr := db.markIndexFailed(ctx, def, desc, err)
	if markErr != nil {
		log.ErrorE("failed to record index failure", errors.Join(markErr, err),
			corelog.String("collectionID", def.CollectionID),
			corelog.Any("indexID", desc.ID),
		)
	}
	return errors.Join(NewErrIndexBackfillFailed(err, desc.Name), markErr)
}

// fillIndexBatches indexes every document from startAfter to the end of the collection in batched
// transactions, returning the epoch it filled. The epoch is resolved once and pinned for the whole
// build.
//
// A non-retryable error marks the index failed; a conflict leaves it resumable. It does not
// complete the build. The caller deletes the record once the fill is done.
func (db *DB) fillIndexBatches(
	ctx context.Context,
	def client.CollectionVersion,
	desc client.IndexDescription,
	startAfter immutable.Option[uint64],
) (uint32, error) {
	fields := make([]client.CollectionFieldDescription, 0, len(desc.Fields))
	for _, f := range desc.Fields {
		if colField, ok := def.GetFieldByName(f.Name); ok {
			fields = append(fields, colField)
		}
	}

	// Resolve the epoch once and pin it for the whole build. A concurrent version switch can advance
	// the sequence mid-build; if each batch re-read it, the build would split across two epochs and
	// leave the new live epoch missing every document indexed before the advance. The version switch
	// stages its own building record, which drives a complete fresh build of the new epoch.
	epoch, err := db.readIndexBuildEpoch(ctx, def.CollectionID, desc.ID)
	if err != nil {
		return 0, err
	}

	watermark := startAfter

	for {
		if IndexBuildGate != nil {
			IndexBuildGate(ctx, def.CollectionID, desc.ID)
		}

		var (
			lastDocShortID uint64
			n              int
		)

		superseded := false
		batchErr := db.withTxnRetries(ctx, func(batchCtx context.Context) error {
			// A version switch mid-build advances the epoch and stages a fresh build to fill it from the
			// top. This build is pinned to the old epoch, so it is now stale: stop, and don't touch the
			// shared record's watermark, which the fresh build resets to rebuild from the start. Reading
			// the epoch here also joins this batch's read-set, so a switch committing after the check but
			// before this batch commits conflicts it (and it retries into this same superseded branch).
			liveEpoch, err := getIndexEpoch(batchCtx, def.CollectionID, desc.ID)
			if err != nil {
				return err
			}
			if liveEpoch != epoch {
				superseded = true
				return nil
			}

			// Bare collection so this batch does not read sibling indexes' state into its read-set
			// (see newBareCollection); the backfill maintains only its own index, built below.
			col := db.newBareCollection(def, datastore.CtxTryGetTxnOption(batchCtx))

			// building=true so Save tolerates entries a concurrent live write
			// already stored for the same document. The epoch is pinned above, not re-read per batch.
			colIndex, err := newCollectionIndexWithEpoch(col, desc, true, epoch)
			if err != nil {
				return err
			}

			lastDocShortID, n, err = col.iterateDocsBatch(
				batchCtx, fields, watermark, indexBackfillBatchSize,
				func(doc *client.Document) error {
					return colIndex.Save(batchCtx, doc)
				},
			)
			if err != nil {
				return err
			}

			// The watermark is only meaningful if the batch processed any documents.
			if n > 0 {
				return db.advanceIndexWatermark(batchCtx, def.CollectionID, desc.ID, lastDocShortID)
			}
			return nil
		})

		if batchErr != nil {
			// A conflicting write or shutdown cancellation leaves the state and watermark valid, so the
			// index can resume. Only a real error marks it failed.
			if isBuildInterrupted(batchErr) {
				return epoch, NewErrIndexBackfillInterrupted(batchErr, desc.Name)
			}
			markErr := db.markIndexFailed(ctx, def, desc, batchErr)
			if markErr != nil {
				log.ErrorE("failed to record index failure",
					errors.Join(markErr, batchErr),
					corelog.String("collectionID", def.CollectionID),
					corelog.Any("indexID", desc.ID),
				)
			}
			return epoch, errors.Join(NewErrIndexBackfillFailed(batchErr, desc.Name), markErr)
		}

		if superseded {
			// The pinned epoch is no longer live; the fresh build staged by the version switch will fill
			// the new epoch. Stop without completing, leaving the record for that build.
			return epoch, nil
		}

		if n == 0 || n < indexBackfillBatchSize {
			break
		}

		watermark = immutable.Some(lastDocShortID)
	}

	return epoch, nil
}

// markIndexFailed records a failed state for the index with the cause as the reason, in its
// own transaction, retrying on transaction conflicts up to db.MaxTxnRetries() times.
func (db *DB) markIndexFailed(
	ctx context.Context,
	def client.CollectionVersion,
	desc client.IndexDescription,
	rootErr error,
) error {
	return db.withTxnRetries(ctx, func(c context.Context) error {
		return db.markIndexBuildFailed(c, def.CollectionID, desc.ID, rootErr.Error())
	})
}

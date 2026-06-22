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
// pack far more docs per transaction and is tracked as a follow-up. It is a var so tests can
// lower it to exercise multi-batch runs.
var indexBackfillBatchSize = 100

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

// backfillIndex builds an index by indexing every existing document in batched transactions, then
// deleting the build record to mark it ready. It is used both for a fresh index and for a rebuild
// filling a new epoch; in both cases the epoch is resolved from the index's sequence.
//
// startAfter resumes after the given docID from a persisted watermark; pass None to build the
// whole collection. A non-retryable error marks the index failed; a conflict leaves it resumable.
func (db *DB) backfillIndex(
	ctx context.Context,
	def client.CollectionVersion,
	desc client.IndexDescription,
	startAfter immutable.Option[string],
) error {
	if err := db.fillIndexBatches(ctx, def, desc, startAfter); err != nil {
		return err
	}

	err := db.withTxnRetries(ctx, func(c context.Context) error {
		return db.completeIndexBuild(c, def.CollectionID, desc.ID)
	})
	if err == nil {
		return nil
	}
	if errors.Is(err, corekv.ErrTxnConflict) {
		return NewErrIndexBackfillInterrupted(err, desc.Name)
	}
	markErr := db.markIndexFailed(ctx, def, desc, err)
	if markErr != nil {
		log.ErrorE("failed to record index failure", markErr,
			corelog.String("collectionID", def.CollectionID),
			corelog.Any("indexID", desc.ID),
		)
	}
	return errors.Join(NewErrIndexBackfillFailed(err, desc.Name), markErr)
}

// fillIndexBatches indexes every document from startAfter to the end of the collection in batched
// transactions. The index writes the epoch resolved from its sequence.
//
// A non-retryable error marks the index failed; a conflict leaves it resumable. It does not
// complete the build — the caller deletes the record once the fill is done.
func (db *DB) fillIndexBatches(
	ctx context.Context,
	def client.CollectionVersion,
	desc client.IndexDescription,
	startAfter immutable.Option[string],
) error {
	fields := make([]client.CollectionFieldDescription, 0, len(desc.Fields))
	for _, f := range desc.Fields {
		if colField, ok := def.GetFieldByName(f.Name); ok {
			fields = append(fields, colField)
		}
	}

	watermark := startAfter

	for {
		var (
			lastDocID string
			n         int
		)

		batchErr := db.withTxnRetries(ctx, func(batchCtx context.Context) error {
			col, err := db.newCollection(batchCtx, def, datastore.CtxTryGetTxnOption(batchCtx))
			if err != nil {
				return err
			}

			// building=true so Save tolerates entries a concurrent live write
			// already stored for the same document.
			colIndex, err := NewCollectionIndex(batchCtx, col, desc, true)
			if err != nil {
				return err
			}

			lastDocID, n, err = col.iterateDocsBatch(
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
				return db.advanceIndexWatermark(batchCtx, def.CollectionID, desc.ID, lastDocID)
			}
			return nil
		})

		if batchErr != nil {
			// A transaction conflict means a concurrent write raced with this batch.
			// The building state and watermark are still valid, so leave the index
			// resumable rather than recording a permanent failure. Only non-retryable
			// errors represent a genuine problem that warrants marking the index failed.
			if errors.Is(batchErr, corekv.ErrTxnConflict) {
				return NewErrIndexBackfillInterrupted(batchErr, desc.Name)
			}
			markErr := db.markIndexFailed(ctx, def, desc, batchErr)
			if markErr != nil {
				log.ErrorE("failed to record index failure",
					markErr,
					corelog.String("collectionID", def.CollectionID),
					corelog.Any("indexID", desc.ID),
				)
			}
			return errors.Join(NewErrIndexBackfillFailed(batchErr, desc.Name), markErr)
		}

		if n == 0 || n < indexBackfillBatchSize {
			break
		}

		watermark = immutable.Some(lastDocID)
	}

	return nil
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

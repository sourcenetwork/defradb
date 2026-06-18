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

// indexBackfillBatchSize is the number of documents indexed per batch transaction
// during backfill. 100 entries per batch stays well under the storage engine's
// transaction size limit. It is a var so tests can lower it to exercise multi-batch runs.
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
		txn := rawTxn.(*Txn)
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

// backfillIndex indexes all existing documents in def for the given index desc,
// running the work in batched transactions so that no single transaction exceeds
// the storage engine's transaction size limit.
//
// startAfter resumes the build after the given docID, used by startup recovery to
// continue an interrupted build from its persisted watermark; pass None to build
// the whole collection.
//
// Batches run concurrently with live writes and are serialized only by the storage
// engine's optimistic conflict detection: a live write to a doc a batch already read
// conflicts at commit, so the batch retries and re-reads the latest state. This is why
// each batch builds a fresh collection and the loop retries on conflict.
func (db *DB) backfillIndex(
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
			colIndex, err := NewCollectionIndex(col, desc, true)
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

	// A missing state record means ready, so completion deletes the record
	// instead of storing a terminal status. Only in-flight and failed
	// indexes keep a record.
	if err := db.withTxnRetries(ctx, func(c context.Context) error {
		return db.deleteIndexState(c, def.CollectionID, desc.ID)
	}); err != nil {
		// A conflict here means entries are all written; state is still building and resumable.
		// Only a non-retryable error warrants marking the index failed.
		if errors.Is(err, corekv.ErrTxnConflict) {
			return NewErrIndexBackfillInterrupted(err, desc.Name)
		}
		markErr := db.markIndexFailed(ctx, def, desc, err)
		if markErr != nil {
			log.ErrorE("failed to record index failure",
				markErr,
				corelog.String("collectionID", def.CollectionID),
				corelog.Any("indexID", desc.ID),
			)
		}
		return errors.Join(NewErrIndexBackfillFailed(err, desc.Name), markErr)
	}
	return nil
}

// setIndexStateWithRetry writes the index state in its own transaction,
// retrying on transaction conflicts up to db.MaxTxnRetries() times.
func (db *DB) setIndexStateWithRetry(
	ctx context.Context,
	collectionID string,
	indexID uint32,
	state indexState,
) error {
	return db.withTxnRetries(ctx, func(c context.Context) error {
		return db.setIndexState(c, collectionID, indexID, state)
	})
}

// markIndexFailed records a failed state for the index with the cause as the reason.
func (db *DB) markIndexFailed(
	ctx context.Context,
	def client.CollectionVersion,
	desc client.IndexDescription,
	rootErr error,
) error {
	return db.setIndexStateWithRetry(ctx, def.CollectionID, desc.ID, indexState{
		Status: client.IndexStatusFailed,
		Reason: rootErr.Error(),
	})
}

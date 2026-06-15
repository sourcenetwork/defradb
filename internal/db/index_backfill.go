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
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

// indexBackfillBatchSize is the number of documents indexed per batch transaction
// during backfill. Index keys embed field values and keys are capped at ~65 KB,
// so 100 docs per batch stays well under the storage engine's transaction size limit.
// It is a var so tests can lower it to exercise multi-batch runs.
var indexBackfillBatchSize = 100

// withTxnRetries runs attempt with a fresh read-write transaction set on the context
// and commits it afterwards. When the attempt or the commit fails with a transaction
// conflict, the attempt is re-run with a new transaction, up to db.MaxTxnRetries() times;
// the last conflict error is returned if retries run out. Other errors abort immediately.
func (db *DB) withTxnRetries(ctx context.Context, attempt func(ctx context.Context) error) error {
	var lastErr error
	for i := 0; i < db.MaxTxnRetries(); i++ {
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
func (db *DB) backfillIndex(
	ctx context.Context,
	def client.CollectionVersion,
	desc client.IndexDescription,
) error {
	fields := make([]client.CollectionFieldDescription, 0, len(desc.Fields))
	for _, f := range desc.Fields {
		if colField, ok := def.GetFieldByName(f.Name); ok {
			fields = append(fields, colField)
		}
	}

	watermark := immutable.None[string]()

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

			colIndex, err := NewCollectionIndex(col, desc)
			if err != nil {
				return err
			}
			// Mark building so Save tolerates entries a concurrent live write
			// already stored for the same document.
			colIndex.setBuilding(true)

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
				return setIndexState(
					batchCtx,
					def.CollectionID,
					desc.ID,
					indexState{Status: client.IndexStatusBuilding, Watermark: lastDocID},
				)
			}
			return nil
		})

		if batchErr != nil {
			markErr := db.markIndexFailed(ctx, def, desc, batchErr)
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
		return deleteIndexState(c, def.CollectionID, desc.ID)
	}); err != nil {
		markErr := db.markIndexFailed(ctx, def, desc, err)
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
		return setIndexState(c, collectionID, indexID, state)
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

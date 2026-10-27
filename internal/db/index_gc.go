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

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// rawBytesKey adapts a raw key (as yielded by an iterator) to datastore.Key for
// deletion. It is not a keys.CollectionedKey, so the datastore skips the
// collection-level lock. It's safe because the definition is already removed,
// so no writer touches these entries while the GC deletes them.
type rawBytesKey struct {
	b []byte
}

func (r rawBytesKey) Bytes() []byte { return r.b }

// gcIndex deletes all of the index's entries, across every epoch, in batched transactions, then
// removes the drop record. Used by a whole-index drop. The caller resolves collectionID and
// collectionShortID while its staging transaction is live.
func (db *DB) gcIndex(
	ctx context.Context,
	collectionID string,
	collectionShortID uint32,
	indexID uint32,
	indexName string,
) error {
	if err := db.gcIndexEntries(ctx, collectionShortID, indexID, indexName); err != nil {
		return err
	}

	if err := db.withTxnRetries(ctx, func(c context.Context) error {
		// Also clear any backfill record. deleteIndex clears it up front, but a build still in flight
		// at that point can re-create one (an advanced watermark, or a failed status) after the clear
		// and before its guard releases the drop. Sweeping it here keeps a deleted index from leaving
		// an orphaned building or failed record behind.
		if err := db.clearIndexBuildRecord(c, collectionID, indexID); err != nil {
			return err
		}
		return db.completeIndexDrop(c, collectionID, indexID)
	}); err != nil {
		return NewErrIndexGCFailed(err, indexName)
	}
	return nil
}

// gcIndexEntries deletes every entry for the index, across every epoch, in batched transactions,
// leaving the state record untouched. Idempotent, so recovery can repeat it.
func (db *DB) gcIndexEntries(ctx context.Context, collectionShortID, indexID uint32, indexName string) error {
	// Epoch 0 carries no component, so this prefix covers every epoch of the index.
	prefixKey := &keys.IndexDataStoreKey{CollectionShortID: collectionShortID, IndexID: indexID}

	for {
		n, batchErr := db.gcIndexBatch(ctx, datastore.IterOptions{Prefix: prefixKey, KeysOnly: true})
		if batchErr != nil {
			return NewErrIndexGCFailed(batchErr, indexName)
		}
		if n < indexBackfillBatchSize {
			break
		}
	}
	return nil
}

// gcStaleEpochs deletes every entry of the index that belongs to a superseded epoch: everything
// from the start of the index keyspace up to, but excluding, liveEpoch. Epochs are allocated by a
// monotonic sequence and encode in ascending order, so all stale epochs sort before the live one
// and a single range delete removes them. Idempotent, so recovery can repeat it.
func (db *DB) gcStaleEpochs(ctx context.Context, collectionShortID, indexID, liveEpoch uint32, indexName string) error {
	start := &keys.IndexDataStoreKey{CollectionShortID: collectionShortID, IndexID: indexID}
	end := &keys.IndexDataStoreKey{CollectionShortID: collectionShortID, IndexID: indexID, Epoch: liveEpoch}

	for {
		n, batchErr := db.gcIndexBatch(ctx, datastore.IterOptions{Start: start, End: end, KeysOnly: true})
		if batchErr != nil {
			return NewErrIndexGCFailed(batchErr, indexName)
		}
		if n < indexBackfillBatchSize {
			break
		}
	}
	return nil
}

// gcIndexBatch deletes up to indexBackfillBatchSize raw index keys matching opts in a single
// committed transaction. It returns the number of keys deleted. A return value smaller than
// indexBackfillBatchSize means no more keys remain. The backfill batch size is reused as the GC
// delete-batch size (single tunable).
func (db *DB) gcIndexBatch(ctx context.Context, opts datastore.IterOptions) (int, error) {
	var n int
	err := db.withTxnRetries(ctx, func(batchCtx context.Context) error {
		n = 0
		txn := datastore.CtxMustGetTxn(batchCtx)

		iter, err := txn.Datastore().Iterator(batchCtx, opts)
		if err != nil {
			return NewErrCreateDeleteIndexIterator(err)
		}

		rawKeys := make([][]byte, 0, indexBackfillBatchSize)
		for len(rawKeys) < indexBackfillBatchSize {
			hasNext, err := iter.Next()
			if err != nil {
				return errors.Join(err, iter.Close())
			}
			if !hasNext {
				break
			}
			rawKeys = append(rawKeys, append([]byte(nil), iter.Key()...))
		}
		if err := iter.Close(); err != nil {
			return err
		}

		for _, k := range rawKeys {
			if err := txn.Datastore().Delete(batchCtx, rawBytesKey{k}); err != nil {
				return err
			}
		}
		n = len(rawKeys)
		return nil
	})
	return n, err
}

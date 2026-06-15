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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/id"
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

// gcIndex deletes all stored entries for the index in batched transactions so that
// no single transaction exceeds the storage engine's transaction size limit, then
// removes the index state record.
func (db *DB) gcIndex(
	ctx context.Context,
	def client.CollectionVersion,
	desc client.IndexDescription,
) error {
	// Resolve the short collection ID inside a transaction, then reuse it
	// across all GC batches (the mapping is immutable once set).
	var shortID uint32
	if err := db.withTxnRetries(ctx, func(txnCtx context.Context) error {
		var resolveErr error
		shortID, resolveErr = id.GetShortCollectionID(txnCtx, def.CollectionID)
		return resolveErr
	}); err != nil {
		return NewErrIndexGCFailed(err, desc.Name)
	}

	prefixKey := &keys.IndexDataStoreKey{
		CollectionShortID: shortID,
		IndexID:           desc.ID,
	}

	for {
		n, batchErr := db.gcIndexBatch(ctx, prefixKey)
		if batchErr != nil {
			return NewErrIndexGCFailed(batchErr, desc.Name)
		}
		if n < indexBackfillBatchSize {
			break
		}
	}

	if err := db.withTxnRetries(ctx, func(c context.Context) error {
		return deleteIndexState(c, def.CollectionID, desc.ID)
	}); err != nil {
		return NewErrIndexGCFailed(err, desc.Name)
	}
	return nil
}

// gcIndexBatch deletes up to indexBackfillBatchSize raw index keys under prefixKey
// in a single committed transaction. It returns the number of keys deleted.
// A return value smaller than indexBackfillBatchSize means no more keys remain.
func (db *DB) gcIndexBatch(ctx context.Context, prefixKey *keys.IndexDataStoreKey) (int, error) {
	var n int
	err := db.withTxnRetries(ctx, func(batchCtx context.Context) error {
		n = 0
		txn := datastore.CtxMustGetTxn(batchCtx)

		iter, err := txn.Datastore().Iterator(batchCtx, datastore.IterOptions{
			Prefix:   prefixKey,
			KeysOnly: true,
		})
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

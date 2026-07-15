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
	"fmt"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/description"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// The functions in this file resolve a single index's transient state: a building record left by an
// interrupted backfill, a dropping record left by interrupted GC, or superseded epochs left by an
// interrupted rebuild. The indexBuildWorker dispatches them, both on startup and on demand for an
// async NewIndex/DeleteIndex. See index_worker.go for the drain loop and concurrency control.
//
// Only index actions are recovered; truncate and datastore refresh are not yet resumed (tracked by
// https://github.com/sourcenetwork/defradb/issues/4874). A half-built index returns incomplete
// query results, so it must be recovered.

// listAllIndexStates opens a read-only transaction, scans all index state records,
// and returns them. The transaction is discarded before returning.
func (db *DB) listAllIndexStates(ctx context.Context) ([]indexStateRecord, error) {
	rawTxn, err := db.NewTxn(true)
	if err != nil {
		return nil, err
	}
	txnCtx := InitContext(ctx, rawTxn)
	defer rawTxn.Discard()

	return listIndexStates(txnCtx)
}

// recoverBuilding resumes an interrupted build from its persisted watermark rather than
// abandoning it. The build fills the epoch the sequence already names, whether it is a fresh
// index or a rebuild's new epoch. A non-retryable error (e.g. a unique-constraint violation) is
// recorded as the failed state by backfillIndex and returned here.
func (db *DB) recoverBuilding(ctx context.Context, key keys.IndexStateKey, state indexState) error {
	def, desc, err := db.findIndexDefinition(ctx, key)
	if err != nil {
		// The record outlived its definition: a crash can leave a building record whose definition
		// was never committed, or a rebuild can orphan one. It can never build, so clear the record
		// rather than return an error the drain would re-dispatch forever.
		if errors.Is(err, ErrIndexWithIDDoesNotExist) {
			return db.withTxnRetries(ctx, func(c context.Context) error {
				return db.clearIndexBuildRecord(c, key.CollectionID, key.IndexID)
			})
		}
		return err
	}

	startAfter := immutable.None[uint64]()
	if state.Watermark != 0 {
		startAfter = immutable.Some(state.Watermark)
	}
	return db.backfillIndex(ctx, def, desc, startAfter)
}

// findIndexDefinition resolves the collection version and index description for the
// given state key from the collection repository. It prefers the active version when
// multiple versions contain the index; if no active version matches, the first match
// is returned. Multiple active versions matching the same index ID should not occur —
// if they do, a warning is logged and the first active match is used.
func (db *DB) findIndexDefinition(
	ctx context.Context,
	key keys.IndexStateKey,
) (client.CollectionVersion, client.IndexDescription, error) {
	rawTxn, err := db.NewTxn(true)
	if err != nil {
		return client.CollectionVersion{}, client.IndexDescription{}, err
	}
	defer rawTxn.Discard()
	txnCtx := InitContext(ctx, rawTxn)

	versions, err := description.GetCollectionsByCollectionID(txnCtx, db.collectionRepository, key.CollectionID)
	if err != nil {
		return client.CollectionVersion{}, client.IndexDescription{}, err
	}

	// Prefer the active version's copy of the index; an index ID can appear on
	// several versions (e.g. across a migration) and the active one serves the
	// live documents this backfill must index.
	found := false
	var matchDef client.CollectionVersion
	var matchIdx client.IndexDescription
	for _, def := range versions {
		for _, idx := range def.Indexes {
			if idx.ID != key.IndexID {
				continue
			}
			if def.IsActive {
				return def, idx, nil
			}
			if !found {
				found, matchDef, matchIdx = true, def, idx
			}
		}
	}

	if found {
		return matchDef, matchIdx, nil
	}
	return client.CollectionVersion{}, client.IndexDescription{},
		NewErrIndexWithIDDoesNotExist(key.IndexID, key.CollectionID)
}

// recoverDropping resumes an interrupted whole-index drop, deleting the remaining entries and the
// drop record. Rebuilds leave no drop record — their superseded epochs are collected by
// recoverStaleEpochs instead.
func (db *DB) recoverDropping(ctx context.Context, key keys.IndexStateKey) error {
	collectionShortID, err := db.resolveCollectionShortID(ctx, key.CollectionID)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("index %d", key.IndexID)
	return db.gcIndex(ctx, key.CollectionID, collectionShortID, key.IndexID, name)
}

// recoverStaleEpochs collects superseded epochs for indexes marked by a rebuild. Only marked indexes
// are swept, so a drain with no pending rebuild is a no-op that touches no storage. Each index keeps
// only its live epoch; everything below it is stale and deleted, then the marker is cleared.
//
// The delete range is bounded strictly below the live epoch, so it never touches an in-progress
// build (which fills the live epoch itself). A superseded epoch holds pre-migration values that no
// query reads — a building index is excluded from planning and full-scans instead — so collecting
// it while a rebuild is still in flight is safe.
func (db *DB) recoverStaleEpochs(ctx context.Context) error {
	markers, err := db.listStaleEpochMarkers(ctx)
	if err != nil {
		return err
	}

	for _, m := range markers {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		liveEpoch, err := db.indexLiveEpoch(ctx, m.CollectionShortID, m.IndexID)
		if err != nil {
			return err
		}
		name := fmt.Sprintf("index %d", m.IndexID)
		if err := db.gcStaleEpochs(ctx, m.CollectionShortID, m.IndexID, liveEpoch, name); err != nil {
			return err
		}
		if err := db.clearStaleEpochMarker(ctx, m.CollectionShortID, m.IndexID); err != nil {
			return err
		}
	}
	return nil
}

// listStaleEpochMarkers reads every stale-epoch marker in a short read-only transaction.
func (db *DB) listStaleEpochMarkers(ctx context.Context) ([]keys.IndexStaleEpochKey, error) {
	rawTxn, err := db.NewTxn(true)
	if err != nil {
		return nil, err
	}
	defer rawTxn.Discard()
	txn := datastore.CtxMustGetTxn(InitContext(ctx, rawTxn))

	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix:   []byte(keys.IndexStaleEpochPrefix()),
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

	var markers []keys.IndexStaleEpochKey
	for {
		hasNext, err := iter.Next()
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}
		k, err := keys.NewIndexStaleEpochKeyFromString(string(iter.Key()))
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}
		markers = append(markers, k)
	}
	return markers, iter.Close()
}

// clearStaleEpochMarker deletes an index's stale-epoch marker once its stale epochs are collected.
func (db *DB) clearStaleEpochMarker(ctx context.Context, collectionShortID, indexID uint32) error {
	return db.withTxnRetries(ctx, func(c context.Context) error {
		txn := datastore.CtxMustGetTxn(c)
		return txn.Systemstore().Delete(c, keys.NewIndexStaleEpochKey(collectionShortID, indexID).Bytes())
	})
}

// markStaleEpochs records that an index has superseded epochs to collect, on the transaction bound
// to ctx so it commits with the rebuild that advanced the epoch. The worker's sweep reads it, GCs
// the stale epochs, and clears it.
func (db *DB) markStaleEpochs(ctx context.Context, collectionShortID, indexID uint32) error {
	txn := datastore.CtxMustGetTxn(ctx)
	return txn.Systemstore().Set(ctx, keys.NewIndexStaleEpochKey(collectionShortID, indexID).Bytes(), []byte{})
}

// indexLiveEpoch reads an index's live epoch (its sequence value) in a short read-only transaction.
func (db *DB) indexLiveEpoch(ctx context.Context, collectionShortID, indexID uint32) (uint32, error) {
	rawTxn, err := db.NewTxn(true)
	if err != nil {
		return 0, err
	}
	defer rawTxn.Discard()
	return readIndexEpochByShortID(InitContext(ctx, rawTxn), collectionShortID, indexID)
}

// resolveCollectionShortID opens a read-only transaction to look up the short
// collection ID, then discards the transaction.
func (db *DB) resolveCollectionShortID(ctx context.Context, collectionID string) (uint32, error) {
	rawTxn, err := db.NewTxn(true)
	if err != nil {
		return 0, err
	}
	defer rawTxn.Discard()
	return id.GetCollectionShortID(InitContext(ctx, rawTxn), collectionID)
}

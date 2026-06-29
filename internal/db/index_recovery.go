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

	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/description"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// recoverIndexStates inspects all index state records and resolves any that were
// left in a transient state by a previous interrupted shutdown.
//
// A building record means a backfill was interrupted; the build is resumed from
// its persisted watermark.
// A dropping record means GC was interrupted; deletion is resumed.
// Failed and ready records are left untouched.
//
// Errors from individual recoveries are logged and skipped so that a partially
// recoverable database can still open.
//
// Indexes are the only action recovered on startup for v1; truncate and datastore refresh
// are not yet resumed (tracked by https://github.com/sourcenetwork/defradb/issues/4874). They
// are recovered because a half-built index returns incomplete query results.
func (db *DB) recoverIndexStates(ctx context.Context) error {
	// Each recovery helper opens its own transaction, so the listing below is read in a
	// separate short-lived transaction that is discarded before any mutation. Recovery thus
	// never holds two transactions at once, keeping it safe on stores that forbid concurrent
	// transactions (leveldb).
	states, err := db.listAllIndexStates(ctx)
	if err != nil {
		log.ErrorE("Failed to list index states during recovery", err)
		return nil
	}

	for _, rec := range states {
		if ctx.Err() != nil {
			return nil
		}
		switch {
		case rec.State.isBuilding():
			if err := db.recoverBuilding(ctx, rec.Key, rec.State); err != nil {
				log.ErrorE("Failed to recover building index", err,
					corelog.String("collectionID", rec.Key.CollectionID),
					corelog.Any("indexID", rec.Key.IndexID),
				)
			}
		case rec.State.isDropping():
			if err := db.recoverDropping(ctx, rec.Key); err != nil {
				log.ErrorE("Failed to recover dropping index", err,
					corelog.String("collectionID", rec.Key.CollectionID),
					corelog.Any("indexID", rec.Key.IndexID),
				)
			}
		default:
			// A failed index requires no recovery action.
		}
	}

	// A rebuild leaves superseded epochs with no record, and may have crashed after its build
	// finished but before collecting them. Sweep every index so any such stale entries are
	// collected; it is a no-op for an index with only its live epoch present.
	if err := db.recoverStaleEpochs(ctx); err != nil {
		log.ErrorE("Failed to collect stale index epochs during recovery", err)
	}
	return nil
}

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

// recoverStaleEpochs collects superseded epochs left by interrupted rebuilds across every active
// index. Each index keeps only its live epoch (the sequence value); everything below it is stale
// and deleted. It is a no-op for an index that has only its live epoch.
//
// The delete range is bounded strictly below the live epoch, so it never touches an in-progress
// build (which fills the live epoch itself). A superseded epoch holds pre-migration values that no
// query reads — a building index is excluded from planning and full-scans instead — so collecting
// it while a rebuild is still in flight is safe.
func (db *DB) recoverStaleEpochs(ctx context.Context) error {
	rawTxn, err := db.NewTxn(true)
	if err != nil {
		return err
	}
	txnCtx := InitContext(ctx, rawTxn)
	cols, err := description.GetActiveCollections(txnCtx, db.collectionRepository)
	rawTxn.Discard()
	if err != nil {
		return err
	}

	for _, col := range cols {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if len(col.Indexes) == 0 {
			continue
		}
		collectionShortID, err := db.resolveCollectionShortID(ctx, col.CollectionID)
		if err != nil {
			return err
		}
		for _, desc := range col.Indexes {
			liveEpoch, err := db.indexLiveEpoch(ctx, col.CollectionID, desc.ID)
			if err != nil {
				return err
			}
			name := fmt.Sprintf("index %d", desc.ID)
			if err := db.gcStaleEpochs(ctx, collectionShortID, desc.ID, liveEpoch, name); err != nil {
				return err
			}
		}
	}
	return nil
}

// indexLiveEpoch reads an index's live epoch (its sequence value) in a short read-only
// transaction.
func (db *DB) indexLiveEpoch(ctx context.Context, collectionID string, indexID uint32) (uint32, error) {
	rawTxn, err := db.NewTxn(true)
	if err != nil {
		return 0, err
	}
	defer rawTxn.Discard()
	return getIndexEpoch(InitContext(ctx, rawTxn), collectionID, indexID)
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

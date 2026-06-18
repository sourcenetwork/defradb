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
func (db *DB) recoverIndexStates(ctx context.Context) error {
	// Each recovery helper opens its own transaction, so the listing below is read
	// in a separate short-lived transaction that is discarded before any mutation.
	states, err := db.listAllIndexStates(ctx)
	if err != nil {
		log.ErrorE("Failed to list index states during recovery", err)
		return nil
	}
	if len(states) == 0 {
		return nil
	}

	for key, state := range states {
		if ctx.Err() != nil {
			return nil
		}
		switch {
		case state.isBuilding():
			if err := db.recoverBuilding(ctx, key, state.Watermark); err != nil {
				log.ErrorE("Failed to recover building index", err,
					corelog.String("collectionID", key.CollectionID),
					corelog.Any("indexID", key.IndexID),
				)
			}
		case state.isDropping():
			if err := db.recoverDropping(ctx, key); err != nil {
				log.ErrorE("Failed to recover dropping index", err,
					corelog.String("collectionID", key.CollectionID),
					corelog.Any("indexID", key.IndexID),
				)
			}
		default:
			// A failed index requires no recovery action.
		}
	}
	return nil
}

// listAllIndexStates opens a read-only transaction, scans all index state records,
// and returns them. The transaction is discarded before returning.
func (db *DB) listAllIndexStates(ctx context.Context) (map[keys.IndexStateKey]indexState, error) {
	rawTxn, err := db.NewTxn(true)
	if err != nil {
		return nil, err
	}
	txnCtx := InitContext(ctx, rawTxn)
	defer rawTxn.Discard()

	return listIndexStates(txnCtx)
}

// recoverBuilding resumes an interrupted backfill from its persisted watermark.
// An interrupted build is not itself a problem with the index, so the build is
// continued rather than abandoned. If the resumed build hits a non-retryable error
// (e.g. the data violates a unique constraint), backfillIndex records the failed
// state itself, so this returns that error without further action.
func (db *DB) recoverBuilding(ctx context.Context, key keys.IndexStateKey, watermark string) error {
	def, desc, err := db.findIndexDefinition(ctx, key)
	if err != nil {
		return err
	}

	startAfter := immutable.None[string]()
	if watermark != "" {
		startAfter = immutable.Some(watermark)
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

// recoverDropping resumes an interrupted GC run for the given index. The short
// collection ID is resolved from the systemstore and gcIndex is called to delete
// the remaining entries and remove the state record.
func (db *DB) recoverDropping(ctx context.Context, key keys.IndexStateKey) error {
	shortID, err := db.resolveShortCollectionID(ctx, key.CollectionID)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("index %d", key.IndexID)
	return db.gcIndex(ctx, key.CollectionID, shortID, key.IndexID, name)
}

// resolveShortCollectionID opens a read-only transaction to look up the short
// collection ID, then discards the transaction.
func (db *DB) resolveShortCollectionID(ctx context.Context, collectionID string) (uint32, error) {
	rawTxn, err := db.NewTxn(true)
	if err != nil {
		return 0, err
	}
	defer rawTxn.Discard()
	return id.GetShortCollectionID(InitContext(ctx, rawTxn), collectionID)
}

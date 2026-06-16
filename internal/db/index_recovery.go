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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// recoverIndexStates inspects all index state records and resolves any that were
// left in a transient state by a previous interrupted shutdown.
//
// A building record means a backfill was interrupted; the index is marked failed.
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
		switch state.Status {
		case client.IndexStatusBuilding:
			if err := db.recoverBuilding(ctx, key); err != nil {
				log.ErrorE("Failed to recover building index", err,
					corelog.String("collectionID", key.CollectionID),
					corelog.Any("indexID", key.IndexID),
				)
			}
		case client.IndexStatusDropping:
			if err := db.recoverDropping(ctx, key); err != nil {
				log.ErrorE("Failed to recover dropping index", err,
					corelog.String("collectionID", key.CollectionID),
					corelog.Any("indexID", key.IndexID),
				)
			}
		default:
			// IndexStatusFailed and IndexStatusReady require no action.
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

// recoverBuilding marks a building index as failed. The backfill was interrupted and
// cannot be safely resumed without tracking where it left off, so the index is put
// into a terminal failed state that the user must resolve by deleting the index.
func (db *DB) recoverBuilding(ctx context.Context, key keys.IndexStateKey) error {
	return db.setIndexStateWithRetry(ctx, key.CollectionID, key.IndexID, indexState{
		Status: client.IndexStatusFailed,
		Reason: "index build interrupted by shutdown",
	})
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

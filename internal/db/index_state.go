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
	"encoding/json"
	"strconv"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/action"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// indexState is the in-memory view of an index's lifecycle action record (see internal/db/action).
// An index has no status of its own: it is an (Action, Status) pair plus the action's reason and
// payload. A missing record means ready.
//
//	building → BackfillIndexAction + InProgress (Watermark set)
//	failed   → BackfillIndexAction + Errored    (Reason set)
//	dropping → DropIndexAction     + InProgress
type indexState struct {
	Action client.Action
	Status client.ActionStatus
	// Reason is the action's generic error reason, set when failed.
	Reason string
	// Watermark is the last indexed docID, decoded from the action payload while building.
	Watermark string
}

// indexPayload is how the index action encodes its watermark into the action's opaque payload.
type indexPayload struct {
	Watermark string
}

func (s indexState) isBuilding() bool {
	return s.Action == client.BackfillIndexAction && s.Status == client.InProgressActionStatus
}

func (s indexState) isFailed() bool {
	return s.Action == client.BackfillIndexAction && s.Status == client.ErroredActionStatus
}

func (s indexState) isDropping() bool {
	return s.Action == client.DropIndexAction && s.Status == client.InProgressActionStatus
}

// statusDescription builds the public status view of the index from its state.
func (s indexState) statusDescription(desc client.IndexDescription, hasState bool) client.IndexDescriptionStatus {
	result := client.IndexDescriptionStatus{IndexDescription: desc}
	if !hasState {
		// No record means ready.
		result.Status = client.CompletedActionStatus
		return result
	}
	result.Status = s.Status
	result.Action = s.Action
	result.Reason = s.Reason
	return result
}

// indexSubject is the action subject segment for an index action: the index ID.
func indexSubject(indexID uint32) string {
	return strconv.FormatUint(uint64(indexID), 10)
}

// isIndexAction reports whether the action describes an index lifecycle state. Action records
// for other actions (truncate, datastore refresh) are ignored by the index state helpers.
func isIndexAction(a client.Action) bool {
	return a == client.BackfillIndexAction || a == client.DropIndexAction
}

// getIndexState retrieves the runtime state for the given index.
//
// If no record describes the index, the returned error satisfies
// errors.Is(err, corekv.ErrNotFound). Callers may treat a missing state as ready.
func getIndexState(ctx context.Context, collectionID string, indexID uint32) (indexState, error) {
	states, err := getIndexStates(ctx, collectionID)
	if err != nil {
		return indexState{}, err
	}
	state, ok := states[indexID]
	if !ok {
		return indexState{}, corekv.ErrNotFound
	}
	return state, nil
}

// startIndexBuild records the start of a backfill and publishes an event. The record is
// written on the transaction bound to ctx, so it commits atomically with any other work on
// that transaction.
func (db *DB) startIndexBuild(ctx context.Context, collectionID string, indexID uint32) error {
	return action.SetTxn(
		ctx, db.events, collectionID, client.BackfillIndexAction, indexSubject(indexID),
		client.InProgressActionStatus, "", nil, true,
	)
}

// startIndexDrop records the start of a drop and publishes an event.
func (db *DB) startIndexDrop(ctx context.Context, collectionID string, indexID uint32) error {
	return action.SetTxn(
		ctx, db.events, collectionID, client.DropIndexAction, indexSubject(indexID),
		client.InProgressActionStatus, "", nil, true,
	)
}

// advanceIndexWatermark records build progress for a building index without publishing an
// event, since the status is unchanged from the initial building transition. The watermark
// rides the transaction bound to ctx so it stays consistent with the index entries written in
// the same batch.
func (db *DB) advanceIndexWatermark(ctx context.Context, collectionID string, indexID uint32, watermark string) error {
	payload, err := json.Marshal(indexPayload{Watermark: watermark})
	if err != nil {
		return err
	}
	return action.SetTxn(
		ctx, db.events, collectionID, client.BackfillIndexAction, indexSubject(indexID),
		client.InProgressActionStatus, "", payload, false,
	)
}

// markIndexBuildFailed records a failed backfill with the cause as the reason.
func (db *DB) markIndexBuildFailed(ctx context.Context, collectionID string, indexID uint32, reason string) error {
	return action.SetTxn(
		ctx, db.events, collectionID, client.BackfillIndexAction, indexSubject(indexID),
		client.ErroredActionStatus, reason, nil, true,
	)
}

// completeIndexBuild deletes the build action record, marking the index ready.
func (db *DB) completeIndexBuild(ctx context.Context, collectionID string, indexID uint32) error {
	return action.CompleteTxn(ctx, db.events, collectionID, client.BackfillIndexAction, indexSubject(indexID))
}

// completeIndexDrop deletes the drop action record, marking the index ready.
func (db *DB) completeIndexDrop(ctx context.Context, collectionID string, indexID uint32) error {
	return action.CompleteTxn(ctx, db.events, collectionID, client.DropIndexAction, indexSubject(indexID))
}

// scanIndexStates scans the action status records under the given prefix and returns the index
// ones, keyed by IndexStateKey.
//
// When skipCorrupt is true an individually unparseable record is logged and skipped rather
// than failing the whole scan, so one bad record does not deny access to an otherwise healthy
// collection. Iterator errors are always fatal.
func scanIndexStates(ctx context.Context, prefix []byte, skipCorrupt bool) (map[keys.IndexStateKey]indexState, error) {
	txn := datastore.CtxMustGetTxn(ctx)

	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix: prefix,
	})
	if err != nil {
		return nil, err
	}

	result := make(map[keys.IndexStateKey]indexState)

	for {
		hasNext, err := iter.Next()
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}

		key := string(iter.Key())

		k, err := keys.NewActionStatusKeyString(key)
		if err != nil {
			if skipCorrupt {
				log.ErrorE("Skipping unparseable action record key", err, corelog.String("key", key))
				continue
			}
			return nil, errors.Join(err, iter.Close())
		}
		// Only index actions with a subject describe an index; skip everything else.
		if k.Subject == "" || !isIndexAction(k.Action) {
			continue
		}

		val, err := iter.Value()
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}

		indexID, err := strconv.ParseUint(k.Subject, 10, 32)
		if err != nil {
			if skipCorrupt {
				log.ErrorE("Skipping index action record with non-numeric subject", err, corelog.String("key", key))
				continue
			}
			return nil, errors.Join(err, iter.Close())
		}

		state, err := loadIndexState(ctx, k, action.DecodeStatus(val))
		if err != nil {
			if skipCorrupt {
				log.ErrorE("Skipping index action record with undecodable data", err, corelog.String("key", key))
				continue
			}
			return nil, errors.Join(err, iter.Close())
		}

		result[keys.IndexStateKey{CollectionID: k.CollectionID, IndexID: uint32(indexID)}] = state
	}

	return result, iter.Close()
}

// loadIndexState builds an indexState for the given action record, fetching the generic reason
// and decoding the index-specific watermark from the opaque action payload.
func loadIndexState(ctx context.Context, k keys.ActionStatusKey, status client.ActionStatus) (indexState, error) {
	reason, err := action.GetReason(ctx, k.CollectionID, k.Action, k.Subject)
	if err != nil {
		return indexState{}, err
	}

	raw, err := action.GetPayload(ctx, k.CollectionID, k.Action, k.Subject)
	if err != nil {
		return indexState{}, err
	}
	var payload indexPayload
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &payload); err != nil {
			return indexState{}, NewErrCorruptIndexPayload(raw)
		}
	}

	return indexState{
		Action:    k.Action,
		Status:    status,
		Reason:    reason,
		Watermark: payload.Watermark,
	}, nil
}

// getIndexStates returns the runtime state for every index belonging to the given collection.
//
// The returned map is keyed by index ID.
func getIndexStates(ctx context.Context, collectionID string) (map[uint32]indexState, error) {
	// Lenient: this feeds collection open and listings, so one corrupt record must not deny
	// access to the whole collection.
	scanned, err := scanIndexStates(ctx, indexActionCollectionPrefix(collectionID), true)
	if err != nil {
		return nil, err
	}

	result := make(map[uint32]indexState, len(scanned))
	for k, v := range scanned {
		result[k.IndexID] = v
	}
	return result, nil
}

// listIndexStates returns the runtime state for every index across all collections.
//
// The returned map is keyed by the full IndexStateKey.
func listIndexStates(ctx context.Context) (map[keys.IndexStateKey]indexState, error) {
	// Strict: recovery should surface a corrupt record rather than silently skip resolving it.
	return scanIndexStates(ctx, keys.NewEmptyActionStatusKey().Bytes(), false)
}

// indexActionCollectionPrefix returns the action status prefix covering every action record
// for the given collection. It is filtered to index actions by scanIndexStates.
func indexActionCollectionPrefix(collectionID string) []byte {
	return keys.NewActionStatusSubjectKey(collectionID, 0, "").CollectionPrefix()
}

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
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// indexState is the in-memory view of an index's lifecycle action record (see internal/db/action).
// An index has no status of its own: it is an (Action, Status) pair plus the action's reason and
// payload. A missing record means ready.
//
//	building  → BackfillIndexAction + InProgress (Watermark set)
//	failed    → BackfillIndexAction + Errored    (Reason set)
//	dropping  → DropIndexAction     + InProgress (whole-index drop)
//
// A rebuild collects the epochs it supersedes without a record; see gcStaleEpochs.
type indexState struct {
	Action client.Action
	Status client.ActionStatus
	// Reason is the action's generic error reason, set when failed.
	Reason string
	// Watermark is the last indexed document short ID, decoded from the action payload while building.
	Watermark uint64
}

// buildPayload is the opaque payload of a build (backfill) action record.
type buildPayload struct {
	// Watermark is the last indexed document short ID, letting an interrupted build resume.
	Watermark uint64
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

// listResult builds the public ListIndexes entry for the index from its state. hasState is
// false for a ready index (no action record), reported as a Completed execution.
func (s indexState) listResult(
	collectionID string,
	desc client.IndexDescription,
	hasState bool,
) client.ListIndexesResult {
	exec := client.ActionExecution{
		CollectionID: collectionID,
		Subject:      indexSubject(desc.ID),
	}
	if !hasState {
		exec.Status = client.CompletedActionStatus
	} else {
		exec.Action = s.Action
		exec.Status = s.Status
		exec.Reason = s.Reason
	}
	return client.ListIndexesResult{Description: desc, Execution: exec}
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

// getIndexState retrieves a runtime action record for the given index (build or drop).
//
// If no record describes the index, the returned error satisfies
// errors.Is(err, corekv.ErrNotFound). Callers may treat a missing state as ready.
func getIndexState(ctx context.Context, collectionID string, indexID uint32) (indexState, error) {
	records, err := scanIndexStates(ctx, indexActionCollectionPrefix(collectionID), true)
	if err != nil {
		return indexState{}, err
	}
	for _, rec := range records {
		if rec.Key.IndexID == indexID {
			return rec.State, nil
		}
	}
	return indexState{}, corekv.ErrNotFound
}

// getIndexEpoch returns the index entry namespace an index reads and writes, resolving it from
// the index's epoch sequence using the transaction bound to ctx. See fetcher.ReadIndexEpoch.
func getIndexEpoch(ctx context.Context, collectionID string, indexID uint32) (uint32, error) {
	return fetcher.ReadIndexEpoch(ctx, datastore.CtxMustGetTxn(ctx), collectionID, indexID)
}

// readIndexEpochByShortID resolves the live epoch from the epoch sequence keyed by the collection's
// short ID, for callers that already have it (the stale-epoch marker stores it).
func readIndexEpochByShortID(ctx context.Context, collectionShortID, indexID uint32) (uint32, error) {
	return fetcher.ReadIndexEpochByShortID(ctx, datastore.CtxMustGetTxn(ctx), collectionShortID, indexID)
}

// readIndexBuildEpoch resolves the epoch a backfill fills, in its own short read-only transaction so
// the caller can pin it for the whole build independent of any batch transaction.
func (db *DB) readIndexBuildEpoch(ctx context.Context, collectionID string, indexID uint32) (uint32, error) {
	rawTxn, err := db.NewTxn(true)
	if err != nil {
		return 0, err
	}
	defer rawTxn.Discard()
	return getIndexEpoch(InitContext(ctx, rawTxn), collectionID, indexID)
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

// startIndexDrop records the start of a whole-index drop and publishes an event.
func (db *DB) startIndexDrop(ctx context.Context, collectionID string, indexID uint32) error {
	return action.SetTxn(
		ctx, db.events, collectionID, client.DropIndexAction, indexSubject(indexID),
		client.InProgressActionStatus, "", nil, true,
	)
}

// advanceIndexWatermark records build progress on the transaction bound to ctx, without
// publishing an event since the status is unchanged.
func (db *DB) advanceIndexWatermark(ctx context.Context, collectionID string, indexID uint32, watermark uint64) error {
	payload, err := json.Marshal(buildPayload{Watermark: watermark})
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

// clearIndexBuildRecord deletes any backfill action record (status, reason, payload) for the index
// without publishing an event. Used when dropping an index to discard a leftover building or failed
// record, which would otherwise be orphaned once the index definition is gone.
func (db *DB) clearIndexBuildRecord(ctx context.Context, collectionID string, indexID uint32) error {
	return action.ClearTxn(ctx, collectionID, client.BackfillIndexAction, indexSubject(indexID))
}

// indexStateRecord is one index action record: its index identity plus the decoded state. An
// index can have more than one (a concurrent build and drop), so records are returned as a slice
// rather than a map keyed by index.
type indexStateRecord struct {
	Key   keys.IndexStateKey
	State indexState
}

// scanIndexStates scans the action status records under the given prefix and returns every index
// one as a separate record.
//
// When skipCorrupt is true an individually unparseable record is logged and skipped rather
// than failing the whole scan, so one bad record does not deny access to an otherwise healthy
// collection. Iterator errors are always fatal.
func scanIndexStates(ctx context.Context, prefix []byte, skipCorrupt bool) ([]indexStateRecord, error) {
	txn := datastore.CtxMustGetTxn(ctx)

	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix: prefix,
	})
	if err != nil {
		return nil, err
	}

	var result []indexStateRecord

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

		status, err := action.DecodeStatus(val)
		if err != nil {
			if skipCorrupt {
				log.ErrorE("Skipping index action record with invalid status encoding", err, corelog.String("key", key))
				continue
			}
			return nil, errors.Join(err, iter.Close())
		}

		state, err := loadIndexState(ctx, k, status)
		if err != nil {
			if skipCorrupt {
				log.ErrorE("Skipping index action record with undecodable data", err, corelog.String("key", key))
				continue
			}
			return nil, errors.Join(err, iter.Close())
		}

		result = append(result, indexStateRecord{
			Key:   keys.IndexStateKey{CollectionID: k.CollectionID, IndexID: uint32(indexID)},
			State: state,
		})
	}

	return result, iter.Close()
}

// loadIndexState builds an indexState for the given action record, fetching the generic reason
// and decoding the action's own payload: a build carries a watermark; a drop carries none.
func loadIndexState(ctx context.Context, k keys.ActionStatusKey, status client.ActionStatus) (indexState, error) {
	reason, err := action.GetReason(ctx, k.CollectionID, k.Action, k.Subject)
	if err != nil {
		return indexState{}, err
	}

	raw, err := action.GetPayload(ctx, k.CollectionID, k.Action, k.Subject)
	if err != nil {
		return indexState{}, err
	}

	state := indexState{Action: k.Action, Status: status, Reason: reason}
	// Only a build record carries a payload, holding its watermark.
	if len(raw) > 0 && k.Action == client.BackfillIndexAction {
		var p buildPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return indexState{}, NewErrCorruptIndexPayload(raw)
		}
		state.Watermark = p.Watermark
	}
	return state, nil
}

// getIndexBuildStates returns the build (backfill) state of every index in the given collection
// that has one, keyed by index ID. It reports only the backfill action, which is what determines
// whether an index is ready, failed or building; a concurrent drop (collecting a superseded epoch)
// is irrelevant to callers of this function and is omitted.
func getIndexBuildStates(ctx context.Context, collectionID string) (map[uint32]indexState, error) {
	// Lenient: this feeds collection open and listings, so one corrupt record must not deny
	// access to the whole collection.
	scanned, err := scanIndexStates(ctx, indexActionCollectionPrefix(collectionID), true)
	if err != nil {
		return nil, err
	}

	result := make(map[uint32]indexState, len(scanned))
	for _, rec := range scanned {
		if rec.State.Action == client.BackfillIndexAction {
			result[rec.Key.IndexID] = rec.State
		}
	}
	return result, nil
}

// listIndexStates returns every index action record across all collections, as a slice so an
// index's concurrent build and drop are both present. Used by recovery, which resumes each
// independently.
func listIndexStates(ctx context.Context) ([]indexStateRecord, error) {
	// Strict: recovery should surface a corrupt record rather than silently skip resolving it.
	return scanIndexStates(ctx, keys.NewEmptyActionStatusKey().Bytes(), false)
}

// indexActionCollectionPrefix returns the action status prefix covering every action record
// for the given collection. It is filtered to index actions by scanIndexStates.
func indexActionCollectionPrefix(collectionID string) []byte {
	return keys.NewActionStatusSubjectKey(collectionID, 0, "").CollectionPrefix()
}

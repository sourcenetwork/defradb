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

// indexState is the in-memory projection of an index's lifecycle action record
// (see internal/db/action): an in-progress BackfillIndexAction is building (Payload
// carries the watermark), an errored one is failed (Reason carries the cause), and an
// in-progress DropIndexAction is dropping. A missing record means ready.
type indexState struct {
	// Status is the current lifecycle state of the index.
	Status client.IndexStatus
	// Watermark is the last indexed docID; set only while building, empty otherwise.
	Watermark string
	// Reason describes the failure; set only when failed, empty otherwise.
	Reason string
}

// indexBackfillPayload is the action record Payload for a building index.
type indexBackfillPayload struct {
	Watermark string `json:"watermark,omitempty"`
}

// indexSubject is the action record subject segment for an index action: the index ID.
func indexSubject(indexID uint32) string {
	return strconv.FormatUint(uint64(indexID), 10)
}

// toIndexState projects a stored action execution onto an indexState.
//
// ok is false for action records that do not describe an index lifecycle state
// (for example truncate or datastore refresh), which callers should ignore.
func toIndexState(exec client.ActionExecution, payload indexBackfillPayload, reason string) (indexState, bool) {
	switch exec.Action {
	case client.BackfillIndexAction:
		switch exec.Status {
		case client.InProgressActionStatus:
			return indexState{Status: client.IndexStatusBuilding, Watermark: payload.Watermark}, true
		case client.ErroredActionStatus:
			return indexState{Status: client.IndexStatusFailed, Reason: reason}, true
		}
	case client.DropIndexAction:
		if exec.Status == client.InProgressActionStatus {
			return indexState{Status: client.IndexStatusDropping}, true
		}
	}
	return indexState{}, false
}

// getIndexState retrieves the runtime state for the given index.
//
// If no record describes the index, the returned error satisfies
// errors.Is(err, corekv.ErrNotFound). Callers may treat a missing state as
// IndexStatusReady.
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

// setIndexState records a lifecycle transition for the index as an action record and
// publishes an ActionExecution event on commit.
//
// The write rides the transaction bound to ctx, committing atomically with any other work
// on that transaction, which keeps a building index's watermark consistent with the index
// entries written in the same batch.
func (db *DB) setIndexState(ctx context.Context, collectionID string, indexID uint32, state indexState) error {
	return db.writeIndexState(ctx, collectionID, indexID, state, true)
}

// advanceIndexWatermark records build progress for a building index without publishing an
// event, since the status is unchanged from the initial building transition.
func (db *DB) advanceIndexWatermark(ctx context.Context, collectionID string, indexID uint32, watermark string) error {
	return db.writeIndexState(
		ctx, collectionID, indexID,
		indexState{Status: client.IndexStatusBuilding, Watermark: watermark},
		false,
	)
}

func (db *DB) writeIndexState(
	ctx context.Context,
	collectionID string,
	indexID uint32,
	state indexState,
	publishEvent bool,
) error {
	subject := indexSubject(indexID)

	switch state.Status {
	case client.IndexStatusBuilding:
		payload, err := json.Marshal(indexBackfillPayload{Watermark: state.Watermark})
		if err != nil {
			return err
		}
		return action.SetTxn(
			ctx, db.events, collectionID, client.BackfillIndexAction, subject,
			client.InProgressActionStatus, "", payload, publishEvent,
		)
	case client.IndexStatusFailed:
		return action.SetTxn(
			ctx, db.events, collectionID, client.BackfillIndexAction, subject,
			client.ErroredActionStatus, state.Reason, nil, publishEvent,
		)
	case client.IndexStatusDropping:
		return action.SetTxn(
			ctx, db.events, collectionID, client.DropIndexAction, subject,
			client.InProgressActionStatus, "", nil, publishEvent,
		)
	default:
		return NewErrInvalidIndexState(state.Status)
	}
}

// completeBackfillState removes the build action record for the index, marking it ready
// (a missing record means ready).
func (db *DB) completeBackfillState(ctx context.Context, collectionID string, indexID uint32) error {
	return action.CompleteTxn(ctx, db.events, collectionID, client.BackfillIndexAction, indexSubject(indexID))
}

// completeDropState removes the drop action record for the index, marking it ready
// (a missing record means ready).
func (db *DB) completeDropState(ctx context.Context, collectionID string, indexID uint32) error {
	return action.CompleteTxn(ctx, db.events, collectionID, client.DropIndexAction, indexSubject(indexID))
}

// scanIndexStates scans the action records reachable under the given key prefix and
// returns those that describe an index lifecycle state, keyed by IndexStateKey.
//
// When skipCorrupt is true an individually unparseable or undecodable record is logged and
// skipped rather than failing the whole scan, so one bad record does not deny access to an
// otherwise healthy collection. Iterator errors are always fatal.
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
		// Only index actions carry a subject; collection-wide actions are skipped.
		if k.Subject == "" {
			continue
		}

		val, err := iter.Value()
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}

		state, ok, err := decodeIndexAction(k, val)
		if err != nil {
			if skipCorrupt {
				log.ErrorE("Skipping undecodable index action record", err, corelog.String("key", key))
				continue
			}
			return nil, errors.Join(err, iter.Close())
		}
		if !ok {
			continue
		}

		indexID, err := strconv.ParseUint(k.Subject, 10, 32)
		if err != nil {
			if skipCorrupt {
				log.ErrorE("Skipping index action record with non-numeric subject", err, corelog.String("key", key))
				continue
			}
			return nil, errors.Join(err, iter.Close())
		}
		result[keys.IndexStateKey{CollectionID: k.CollectionID, IndexID: uint32(indexID)}] = state
	}

	return result, iter.Close()
}

// decodeIndexAction decodes a single action record into an indexState, returning ok=false
// when the record does not describe an index lifecycle state.
func decodeIndexAction(k keys.ActionStatusKey, val []byte) (indexState, bool, error) {
	exec := client.ActionExecution{Action: k.Action}

	env, err := action.DecodeEnvelope(val)
	if err != nil {
		return indexState{}, false, err
	}
	exec.Status = env.Status

	var payload indexBackfillPayload
	if len(env.Payload) > 0 {
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return indexState{}, false, err
		}
	}

	state, ok := toIndexState(exec, payload, env.Reason)
	return state, ok, nil
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

// indexActionCollectionPrefix returns the systemstore prefix covering every action record
// for the given collection. It is filtered to index actions by scanIndexStates.
func indexActionCollectionPrefix(collectionID string) []byte {
	return keys.NewActionStatusSubjectKey(collectionID, 0, "").CollectionPrefix()
}

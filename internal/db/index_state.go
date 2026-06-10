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

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// indexState is the mutable runtime state of an index, stored in the systemstore
// separately from the immutable index description.
type indexState struct {
	// Status is the current lifecycle state of the index.
	Status client.IndexStatus
	// Watermark is the last indexed docID, set while the index is building.
	Watermark string
	// Reason is set when Status is failed, providing a description of the failure.
	Reason string
}

// getIndexState retrieves the runtime state for the given index from the systemstore.
//
// If no state has been stored, the returned error will satisfy errors.Is(err, corekv.ErrNotFound).
// Callers may treat a missing state as IndexStatusReady.
func getIndexState(ctx context.Context, collectionID string, indexID uint32) (indexState, error) {
	txn := datastore.CtxMustGetTxn(ctx)
	key := keys.NewIndexStateKey(collectionID, indexID)

	val, err := txn.Systemstore().Get(ctx, key.Bytes())
	if err != nil {
		return indexState{}, err
	}

	var state indexState
	if err := json.Unmarshal(val, &state); err != nil {
		return indexState{}, err
	}

	return state, nil
}

// setIndexState persists the given runtime state for the index in the systemstore.
func setIndexState(ctx context.Context, collectionID string, indexID uint32, state indexState) error {
	txn := datastore.CtxMustGetTxn(ctx)
	key := keys.NewIndexStateKey(collectionID, indexID)

	buf, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return txn.Systemstore().Set(ctx, key.Bytes(), buf)
}

// deleteIndexState removes the runtime state for the given index from the systemstore.
func deleteIndexState(ctx context.Context, collectionID string, indexID uint32) error {
	txn := datastore.CtxMustGetTxn(ctx)
	key := keys.NewIndexStateKey(collectionID, indexID)
	return txn.Systemstore().Delete(ctx, key.Bytes())
}

// scanIndexStates performs a prefix scan over the systemstore and returns all index state
// entries whose keys start with the given prefix.
//
// The returned map is keyed by the full IndexStateKey.
func scanIndexStates(ctx context.Context, prefix []byte) (map[keys.IndexStateKey]indexState, error) {
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

		k, err := keys.NewIndexStateKeyFromString(string(iter.Key()))
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}

		val, err := iter.Value()
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}

		var state indexState
		if err := json.Unmarshal(val, &state); err != nil {
			return nil, errors.Join(err, iter.Close())
		}

		result[k] = state
	}

	return result, iter.Close()
}

// getIndexStates returns the runtime state for every index belonging to the given collection.
//
// The returned map is keyed by index ID.
func getIndexStates(ctx context.Context, collectionID string) (map[uint32]indexState, error) {
	scanned, err := scanIndexStates(ctx, keys.NewIndexStateCollectionPrefix(collectionID))
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
	return scanIndexStates(ctx, keys.NewIndexStateKeyPrefix())
}

// Copyright 2022 Democratized Data Foundation
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

	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/errors"
)

// AddP2PCollection adds a P2P collection to the stored list
func (db *db) AddP2PCollection(ctx context.Context, collectionID string) error {
	collections, err := db.getP2PCollections(ctx)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}

	isNew := true
	for _, col := range collections {
		if col == collectionID {
			isNew = false
			break
		}
	}

	if isNew {
		collections = append(collections, collectionID)
		colBytes, err := json.Marshal(collections)
		if err != nil {
			return err
		}
		return db.systemstore().Put(ctx, ds.NewKey(core.P2P_COLLECTIONS), colBytes)
	}

	return nil
}

// RemoveP2PCollection removes P2P collection from the stored list
func (db *db) RemoveP2PCollection(ctx context.Context, collectionID string) error {
	collections, err := db.getP2PCollections(ctx)
	if err != nil {
		return err
	}

	found := false
	for i, col := range collections {
		if col == collectionID {
			collections = append(collections[:i], collections[i+1:]...)
			found = true
			break
		}
	}

	if found {
		colBytes, err := json.Marshal(collections)
		if err != nil {
			return err
		}
		return db.systemstore().Put(ctx, ds.NewKey(core.P2P_COLLECTIONS), colBytes)
	}

	return nil
}

// GetAllP2PCollections returns the full list of P2P collections
func (db *db) GetAllP2PCollections(ctx context.Context) ([]string, error) {
	collections, err := db.getP2PCollections(ctx)
	if err != nil {
		return nil, err
	}

	return collections, nil
}

func (db *db) getP2PCollections(ctx context.Context) ([]string, error) {
	value, err := db.systemstore().Get(ctx, ds.NewKey(core.P2P_COLLECTIONS))
	if err != nil {
		return nil, err
	}
	var collections []string
	err = json.Unmarshal(value, &collections)
	if err != nil {
		return nil, err
	}

	return collections, nil
}

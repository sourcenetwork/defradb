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
	"strings"

	jsonpatch "github.com/evanphx/json-patch/v5"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
)

// AddSchema takes the provided schema in SDL format, and applies it to the database,
// and creates the necessary collections, request types, etc.
func (db *db) AddSchema(ctx context.Context, schemaString string) error {
	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	existingDescriptions, err := db.getCollectionDescriptions(ctx, txn)
	if err != nil {
		return err
	}

	newDescriptions, err := db.parser.ParseSDL(ctx, schemaString)
	if err != nil {
		return err
	}

	err = db.parser.SetSchema(ctx, txn, append(existingDescriptions, newDescriptions...))
	if err != nil {
		return err
	}

	for _, desc := range newDescriptions {
		if _, err := db.createCollectionTxn(ctx, txn, desc); err != nil {
			return err
		}
	}

	return txn.Commit(ctx)
}

func (db *db) loadSchema(ctx context.Context, txn datastore.Txn) error {
	descriptions, err := db.getCollectionDescriptions(ctx, txn)
	if err != nil {
		return err
	}

	return db.parser.SetSchema(ctx, txn, descriptions)
}

func (db *db) getCollectionDescriptions(
	ctx context.Context,
	txn datastore.Txn,
) ([]client.CollectionDescription, error) {
	collections, err := db.getAllCollectionsTxn(ctx, txn)
	if err != nil {
		return nil, err
	}

	descriptions := make([]client.CollectionDescription, len(collections))
	for i, collection := range collections {
		descriptions[i] = collection.Description()
	}

	return descriptions, nil
}

// PatchSchema takes the given JSON patch string and applies it to the set of CollectionDescriptions
// present in the database.
//
// It will also update the GQL types used by the query system. It will error and not apply any of the
// requested, valid updates should the net result of the patch result in an invalid state.  The
// individual operations defined in the patch do not need to result in a valid state, only the net result
// of the full patch.
//
// The collections (including the schema version ID) will only be updated if any changes have actually
// been made, if the net result of the patch matches the current persisted description then no changes
// will be applied.
func (db *db) PatchSchema(ctx context.Context, patchString string) error {
	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	patch, err := jsonpatch.DecodePatch([]byte(patchString))
	if err != nil {
		return err
	}

	collectionsByName, err := db.getCollectionsByName(ctx, txn)
	if err != nil {
		return err
	}

	existingDescriptionJson, err := json.Marshal(collectionsByName)
	if err != nil {
		return err
	}

	newDescriptionJson, err := patch.Apply(existingDescriptionJson)
	if err != nil {
		return err
	}

	var newDescriptionsByName map[string]client.CollectionDescription
	decoder := json.NewDecoder(strings.NewReader(string(newDescriptionJson)))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&newDescriptionsByName)
	if err != nil {
		return err
	}

	newDescriptions := []client.CollectionDescription{}
	for _, desc := range newDescriptionsByName {
		newDescriptions = append(newDescriptions, desc)
	}

	for _, desc := range newDescriptions {
		if _, err := db.updateCollectionTxn(ctx, txn, desc); err != nil {
			return err
		}
	}

	err = db.parser.SetSchema(ctx, txn, newDescriptions)
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

func (db *innerDB) getCollectionsByName(
	ctx context.Context,
	txn datastore.Txn,
) (map[string]client.CollectionDescription, error) {
	collections, err := db.getAllCollectionsTxn(ctx, txn)
	if err != nil {
		return nil, err
	}

	collectionsByName := map[string]client.CollectionDescription{}
	for _, collection := range collections {
		collectionsByName[collection.Name()] = collection.Description()
	}

	return collectionsByName, nil
}

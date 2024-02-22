// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package description

import (
	"context"
	"encoding/json"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
)

// SaveCollection saves the given collection to the system store overwriting any
// pre-existing values.
func SaveCollection(
	ctx context.Context,
	txn datastore.Txn,
	desc client.CollectionDescription,
) (client.CollectionDescription, error) {
	buf, err := json.Marshal(desc)
	if err != nil {
		return client.CollectionDescription{}, err
	}

	key := core.NewCollectionKey(desc.ID)
	err = txn.Systemstore().Set(ctx, key.ToDS().Bytes(), buf)
	if err != nil {
		return client.CollectionDescription{}, err
	}

	idBuf, err := json.Marshal(desc.ID)
	if err != nil {
		return client.CollectionDescription{}, err
	}

	nameKey := core.NewCollectionNameKey(desc.Name)
	err = txn.Systemstore().Set(ctx, nameKey.ToDS().Bytes(), idBuf)
	if err != nil {
		return client.CollectionDescription{}, err
	}

	// The need for this key is temporary, we should replace it with the global collection ID
	// https://github.com/sourcenetwork/defradb/issues/1085
	schemaVersionKey := core.NewCollectionSchemaVersionKey(desc.SchemaVersionID, desc.ID)
	err = txn.Systemstore().Set(ctx, schemaVersionKey.ToDS().Bytes(), []byte{})
	if err != nil {
		return client.CollectionDescription{}, err
	}

	return desc, nil
}

// GetCollectionByName returns the collection with the given name.
//
// If no collection of that name is found, it will return an error.
func GetCollectionByName(
	ctx context.Context,
	txn datastore.Txn,
	name string,
) (client.CollectionDescription, error) {
	nameKey := core.NewCollectionNameKey(name)
	idBuf, err := txn.Systemstore().Get(ctx, nameKey.ToDS().Bytes())
	if err != nil {
		return client.CollectionDescription{}, err
	}

	var id uint32
	err = json.Unmarshal(idBuf, &id)
	if err != nil {
		return client.CollectionDescription{}, err
	}

	key := core.NewCollectionKey(id)
	buf, err := txn.Systemstore().Get(ctx, key.ToDS().Bytes())
	if err != nil {
		return client.CollectionDescription{}, err
	}

	var col client.CollectionDescription
	err = json.Unmarshal(buf, &col)
	if err != nil {
		return client.CollectionDescription{}, err
	}

	return col, nil
}

// GetCollectionsBySchemaVersionID returns all collections that use the given
// schemaVersionID.
//
// If no collections are found an empty set will be returned.
func GetCollectionsBySchemaVersionID(
	ctx context.Context,
	txn datastore.Txn,
	schemaVersionID string,
) ([]client.CollectionDescription, error) {
	schemaVersionKey := core.NewCollectionSchemaVersionKey(schemaVersionID, 0)
	schemaVersionIter := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix:   schemaVersionKey.Bytes(),
		KeysOnly: true,
	})

	colIDs := make([]uint32, 0)
	for ; schemaVersionIter.Valid(); schemaVersionIter.Next() {
		keyBuf := string(schemaVersionIter.Key())
		colSchemaVersionKey, err := core.NewCollectionSchemaVersionKeyFromString(keyBuf)
		if err != nil {
			if err := schemaVersionIter.Close(ctx); err != nil {
				return nil, NewErrFailedToCloseSchemaQuery(err)
			}
			return nil, err
		}

		colIDs = append(colIDs, colSchemaVersionKey.CollectionID)
	}
	err := schemaVersionIter.Close(ctx)
	if err != nil {
		return nil, NewErrFailedToCloseSchemaQuery(err)
	}

	cols := make([]client.CollectionDescription, len(colIDs))
	for i, colID := range colIDs {
		key := core.NewCollectionKey(colID)
		buf, err := txn.Systemstore().Get(ctx, key.ToDS().Bytes())
		if err != nil {
			return nil, err
		}

		var col client.CollectionDescription
		err = json.Unmarshal(buf, &col)
		if err != nil {
			return nil, err
		}

		cols[i] = col
	}

	return cols, nil
}

// GetCollectionsBySchemaRoot returns all collections that use the given
// schema root.
//
// If no collections are found an empty set will be returned.
func GetCollectionsBySchemaRoot(
	ctx context.Context,
	txn datastore.Txn,
	schemaRoot string,
) ([]client.CollectionDescription, error) {
	schemaVersionIDs, err := GetSchemaVersionIDs(ctx, txn, schemaRoot)
	if err != nil {
		return nil, err
	}

	cols := []client.CollectionDescription{}
	for _, schemaVersionID := range schemaVersionIDs {
		versionCols, err := GetCollectionsBySchemaVersionID(ctx, txn, schemaVersionID)
		if err != nil {
			return nil, err
		}

		cols = append(cols, versionCols...)
	}

	return cols, nil
}

// GetCollections returns all collections in the system.
func GetCollections(
	ctx context.Context,
	txn datastore.Txn,
) ([]client.CollectionDescription, error) {
	iter := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix: []byte(core.COLLECTION),
	})

	cols := make([]client.CollectionDescription, 0)
	for ; iter.Valid(); iter.Next() {
		var col client.CollectionDescription
		err := json.Unmarshal(iter.Value(), &col)
		if err != nil {
			if err := iter.Close(ctx); err != nil {
				return nil, NewErrFailedToCloseCollectionQuery(err)
			}
			return nil, err
		}

		cols = append(cols, col)
	}

	return cols, nil
}

// HasCollectionByName returns true if there is a collection of the given name,
// else returns false.
func HasCollectionByName(
	ctx context.Context,
	txn datastore.Txn,
	name string,
) (bool, error) {
	nameKey := core.NewCollectionNameKey(name)
	return txn.Systemstore().Has(ctx, nameKey.ToDS().Bytes())
}

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

	"github.com/ipfs/go-datastore/query"

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
	err = txn.Systemstore().Put(ctx, key.ToDS(), buf)
	if err != nil {
		return client.CollectionDescription{}, err
	}

	idBuf, err := json.Marshal(desc.ID)
	if err != nil {
		return client.CollectionDescription{}, err
	}

	nameKey := core.NewCollectionNameKey(desc.Name)
	err = txn.Systemstore().Put(ctx, nameKey.ToDS(), idBuf)
	if err != nil {
		return client.CollectionDescription{}, err
	}

	// The need for this key is temporary, we should replace it with the global collection ID
	// https://github.com/sourcenetwork/defradb/issues/1085
	schemaVersionKey := core.NewCollectionSchemaVersionKey(desc.SchemaVersionID, desc.ID)
	err = txn.Systemstore().Put(ctx, schemaVersionKey.ToDS(), []byte{})
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
	idBuf, err := txn.Systemstore().Get(ctx, nameKey.ToDS())
	if err != nil {
		return client.CollectionDescription{}, err
	}

	var id uint32
	err = json.Unmarshal(idBuf, &id)
	if err != nil {
		return client.CollectionDescription{}, err
	}

	key := core.NewCollectionKey(id)
	buf, err := txn.Systemstore().Get(ctx, key.ToDS())
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

	schemaVersionQuery, err := txn.Systemstore().Query(ctx, query.Query{
		Prefix:   schemaVersionKey.ToString(),
		KeysOnly: true,
	})
	if err != nil {
		return nil, NewErrFailedToCreateCollectionQuery(err)
	}

	colIDs := make([]uint32, 0)
	for res := range schemaVersionQuery.Next() {
		if res.Error != nil {
			if err := schemaVersionQuery.Close(); err != nil {
				return nil, NewErrFailedToCloseSchemaQuery(err)
			}
			return nil, err
		}

		colSchemaVersionKey, err := core.NewCollectionSchemaVersionKeyFromString(string(res.Key))
		if err != nil {
			if err := schemaVersionQuery.Close(); err != nil {
				return nil, NewErrFailedToCloseSchemaQuery(err)
			}
			return nil, err
		}

		colIDs = append(colIDs, colSchemaVersionKey.CollectionID)
	}

	cols := make([]client.CollectionDescription, len(colIDs))
	for i, colID := range colIDs {
		key := core.NewCollectionKey(colID)
		buf, err := txn.Systemstore().Get(ctx, key.ToDS())
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

// GetCollectionsBySchemaID returns all collections that use the given
// schemaID.
//
// If no collections are found an empty set will be returned.
func GetCollectionsBySchemaID(
	ctx context.Context,
	txn datastore.Txn,
	schemaID string,
) ([]client.CollectionDescription, error) {
	schemaVersionIDs, err := GetSchemaVersionIDs(ctx, txn, schemaID)
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
	q, err := txn.Systemstore().Query(ctx, query.Query{
		Prefix: core.COLLECTION,
	})
	if err != nil {
		return nil, NewErrFailedToCreateCollectionQuery(err)
	}

	cols := make([]client.CollectionDescription, 0)
	for res := range q.Next() {
		if res.Error != nil {
			if err := q.Close(); err != nil {
				return nil, NewErrFailedToCloseCollectionQuery(err)
			}
			return nil, err
		}

		var col client.CollectionDescription
		err = json.Unmarshal(res.Value, &col)
		if err != nil {
			if err := q.Close(); err != nil {
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
	return txn.Systemstore().Has(ctx, nameKey.ToDS())
}

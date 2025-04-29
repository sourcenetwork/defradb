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
	"sort"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// SaveCollection saves the given collection to the system store overwriting any
// pre-existing values.
func SaveCollection(
	ctx context.Context,
	txn datastore.Txn,
	desc client.CollectionDescription,
) (client.CollectionDescription, error) {
	if desc.CollectionID != "" {
		// Set the collection short id
		err := id.SetShortCollectionID(ctx, txn, desc.CollectionID)
		if err != nil {
			return client.CollectionDescription{}, err
		}
	}

	existing, err := GetCollectionByID(ctx, txn, desc.ID)
	if err != nil && !errors.Is(err, corekv.ErrNotFound) {
		return client.CollectionDescription{}, err
	}

	buf, err := json.Marshal(desc)
	if err != nil {
		return client.CollectionDescription{}, err
	}

	key := keys.NewCollectionKey(desc.ID)
	err = txn.Systemstore().Set(ctx, key.Bytes(), buf)
	if err != nil {
		return client.CollectionDescription{}, err
	}

	if existing.Name.HasValue() && existing.Name != desc.Name {
		nameKey := keys.NewCollectionNameKey(existing.Name.Value())
		idBuf, err := txn.Systemstore().Get(ctx, nameKey.Bytes())
		nameIndexExsts := true
		if err != nil {
			if errors.Is(err, corekv.ErrNotFound) {
				nameIndexExsts = false
			} else {
				return client.CollectionDescription{}, err
			}
		}
		if nameIndexExsts {
			if string(idBuf) == desc.ID {
				// The name index may have already been overwritten, pointing at another collection
				// we should only remove the existing index if it still points at this collection
				err := txn.Systemstore().Delete(ctx, nameKey.Bytes())
				if err != nil {
					return client.CollectionDescription{}, err
				}
			}
		}
	}

	if desc.Name.HasValue() {
		nameKey := keys.NewCollectionNameKey(desc.Name.Value())
		err = txn.Systemstore().Set(ctx, nameKey.Bytes(), []byte(desc.ID))
		if err != nil {
			return client.CollectionDescription{}, err
		}
	}

	return desc, nil
}

func GetCollectionByID(
	ctx context.Context,
	txn datastore.Txn,
	id string,
) (client.CollectionDescription, error) {
	key := keys.NewCollectionKey(id)
	buf, err := txn.Systemstore().Get(ctx, key.Bytes())
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

// GetCollectionByName returns the collection with the given name.
//
// If no collection of that name is found, it will return an error.
func GetCollectionByName(
	ctx context.Context,
	txn datastore.Txn,
	name string,
) (client.CollectionDescription, error) {
	nameKey := keys.NewCollectionNameKey(name)
	idBuf, err := txn.Systemstore().Get(ctx, nameKey.Bytes())
	if err != nil {
		return client.CollectionDescription{}, err
	}

	return GetCollectionByID(ctx, txn, string(idBuf))
}

// GetCollectionsByCollectionID returns all collection versions for the given id.
//
// If no collections are found an empty set will be returned.
func GetCollectionsByCollectionID(
	ctx context.Context,
	txn datastore.Txn,
	collectionID string,
) ([]client.CollectionDescription, error) { //todo - this should not be dependent on matching to schema root?
	schemaVersionIDs, err := GetSchemaVersionIDs(ctx, txn, collectionID)
	if err != nil {
		return nil, err
	}

	cols := []client.CollectionDescription{}
	for _, schemaVersionID := range schemaVersionIDs {
		versionCol, err := GetCollectionByID(ctx, txn, schemaVersionID)
		if err != nil {
			if errors.Is(err, corekv.ErrNotFound) {
				continue
			}
			return nil, err
		}

		cols = append(cols, versionCol)
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
		versionCol, err := GetCollectionByID(ctx, txn, schemaVersionID)
		if err != nil {
			if errors.Is(err, corekv.ErrNotFound) {
				continue
			}
			return nil, err
		}

		cols = append(cols, versionCol)
	}

	return cols, nil
}

// GetCollections returns all collections in the system.
//
// This includes inactive collections.
func GetCollections(
	ctx context.Context,
	txn datastore.Txn,
) ([]client.CollectionDescription, error) {
	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix: []byte(keys.COLLECTION_ID),
	})
	if err != nil {
		return nil, err
	}

	cols := make([]client.CollectionDescription, 0)
	for {
		hasValue, err := iter.Next()
		if err != nil {
			if err := iter.Close(); err != nil {
				return nil, NewErrFailedToCloseCollectionQuery(err)
			}
			return nil, err
		}

		if !hasValue {
			break
		}

		value, err := iter.Value()
		if err != nil {
			if err := iter.Close(); err != nil {
				return nil, NewErrFailedToCloseCollectionQuery(err)
			}
			return nil, err
		}

		var col client.CollectionDescription
		err = json.Unmarshal(value, &col)
		if err != nil {
			if err := iter.Close(); err != nil {
				return nil, NewErrFailedToCloseCollectionQuery(err)
			}
			return nil, err
		}

		cols = append(cols, col)
	}

	return cols, iter.Close()
}

// GetActiveCollections returns all active collections in the system.
func GetActiveCollections(
	ctx context.Context,
	txn datastore.Txn,
) ([]client.CollectionDescription, error) {
	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix: keys.NewCollectionNameKey("").Bytes(),
	})
	if err != nil {
		return nil, err
	}

	cols := make([]client.CollectionDescription, 0)
	for {
		hasValue, err := iter.Next()
		if err != nil {
			if err := iter.Close(); err != nil {
				return nil, NewErrFailedToCloseCollectionQuery(err)
			}
			return nil, err
		}

		if !hasValue {
			break
		}

		value, err := iter.Value()
		if err != nil {
			if err := iter.Close(); err != nil {
				return nil, NewErrFailedToCloseCollectionQuery(err)
			}
			return nil, err
		}

		col, err := GetCollectionByID(ctx, txn, string(value))
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}

		cols = append(cols, col)
	}

	// Sort the results by ID, so that the order matches that of [GetCollections].
	sort.Slice(cols, func(i, j int) bool { return cols[i].ID < cols[j].ID })

	return cols, iter.Close()
}

// HasCollectionByName returns true if there is a collection of the given name,
// else returns false.
func HasCollectionByName(
	ctx context.Context,
	txn datastore.Txn,
	name string,
) (bool, error) {
	nameKey := keys.NewCollectionNameKey(name)
	return txn.Systemstore().Has(ctx, nameKey.Bytes())
}

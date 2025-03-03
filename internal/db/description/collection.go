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
	"github.com/sourcenetwork/defradb/internal/keys"
)

// SaveCollection saves the given collection to the system store overwriting any
// pre-existing values.
func SaveCollection(
	ctx context.Context,
	txn datastore.Txn,
	desc client.CollectionDescription,
) (client.CollectionDescription, error) {
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
			var keyID uint32
			err = json.Unmarshal(idBuf, &keyID)
			if err != nil {
				return client.CollectionDescription{}, err
			}

			if keyID == desc.ID {
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
		idBuf, err := json.Marshal(desc.ID)
		if err != nil {
			return client.CollectionDescription{}, err
		}

		nameKey := keys.NewCollectionNameKey(desc.Name.Value())
		err = txn.Systemstore().Set(ctx, nameKey.Bytes(), idBuf)
		if err != nil {
			return client.CollectionDescription{}, err
		}
	}

	// The need for this key is temporary, we should replace it with the global collection ID
	// https://github.com/sourcenetwork/defradb/issues/1085
	schemaVersionKey := keys.NewCollectionSchemaVersionKey(desc.SchemaVersionID, desc.ID)
	err = txn.Systemstore().Set(ctx, schemaVersionKey.Bytes(), []byte{})
	if err != nil {
		return client.CollectionDescription{}, err
	}

	rootKey := keys.NewCollectionRootKey(desc.RootID, desc.ID)
	err = txn.Systemstore().Set(ctx, rootKey.Bytes(), []byte{})
	if err != nil {
		return client.CollectionDescription{}, err
	}

	return desc, nil
}

func GetCollectionByID(
	ctx context.Context,
	txn datastore.Txn,
	id uint32,
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

	var id uint32
	err = json.Unmarshal(idBuf, &id)
	if err != nil {
		return client.CollectionDescription{}, err
	}

	return GetCollectionByID(ctx, txn, id)
}

func GetCollectionsByRoot(
	ctx context.Context,
	txn datastore.Txn,
	root uint32,
) ([]client.CollectionDescription, error) {
	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix:   keys.NewCollectionRootKey(root, 0).Bytes(),
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

	cols := []client.CollectionDescription{}
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

		rootKey, err := keys.NewCollectionRootKeyFromString(string(iter.Key()))
		if err != nil {
			if err := iter.Close(); err != nil {
				return nil, NewErrFailedToCloseCollectionQuery(err)
			}
			return nil, err
		}

		col, err := GetCollectionByID(ctx, txn, rootKey.CollectionID)
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}

		cols = append(cols, col)
	}

	return cols, iter.Close()
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
	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix:   keys.NewCollectionSchemaVersionKey(schemaVersionID, 0).Bytes(),
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

	colIDs := make([]uint32, 0)
	for {
		hasValue, err := iter.Next()
		if err != nil {
			if err := iter.Close(); err != nil {
				return nil, NewErrFailedToCloseSchemaQuery(err)
			}
			return nil, err
		}

		if !hasValue {
			break
		}

		colSchemaVersionKey, err := keys.NewCollectionSchemaVersionKeyFromString(string(iter.Key()))
		if err != nil {
			if err := iter.Close(); err != nil {
				return nil, NewErrFailedToCloseSchemaQuery(err)
			}
			return nil, err
		}

		colIDs = append(colIDs, colSchemaVersionKey.CollectionID)
	}

	err = iter.Close()
	if err != nil {
		return nil, err
	}

	cols := make([]client.CollectionDescription, len(colIDs))
	for i, colID := range colIDs {
		key := keys.NewCollectionKey(colID)
		buf, err := txn.Systemstore().Get(ctx, key.Bytes())
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

		var id uint32
		err = json.Unmarshal(value, &id)
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}

		col, err := GetCollectionByID(ctx, txn, id)
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

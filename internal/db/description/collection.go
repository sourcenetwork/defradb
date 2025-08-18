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
	"fmt"
	"sort"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// SaveCollection saves the given collection to the system store.
func SaveCollection(
	ctx context.Context,
	desc client.CollectionVersion,
) error {
	txn := datastore.CtxMustGetTxn(ctx)

	err := id.SetShortCollectionID(ctx, desc.CollectionID)
	if err != nil {
		return err
	}

	err = id.SetShortFieldIDs(ctx, desc)
	if err != nil {
		return err
	}

	buf, err := json.Marshal(desc)
	if err != nil {
		return err
	}

	key := keys.NewCollectionKey(desc.VersionID)
	err = txn.Systemstore().Set(ctx, key.Bytes(), buf)
	if err != nil {
		return err
	}

	if !desc.IsActive {
		nameKey := keys.NewCollectionNameKey(desc.Name)
		idBytes, err := txn.Systemstore().Get(ctx, nameKey.Bytes())
		if err != nil {
			if !errors.Is(err, corekv.ErrNotFound) {
				return err
			}
		}

		if string(idBytes) == desc.VersionID {
			err := txn.Systemstore().Delete(ctx, nameKey.Bytes())
			if err != nil {
				return err
			}
		}
	}

	if desc.IsActive {
		nameKey := keys.NewCollectionNameKey(desc.Name)
		err = txn.Systemstore().Set(ctx, nameKey.Bytes(), []byte(desc.VersionID))
		if err != nil {
			return err
		}
	}

	isNew := desc.CollectionID == desc.VersionID
	if !isNew {
		// We don't need to index the version by collection id, if the version id is the collection id
		collectionVersionKey := keys.NewCollectionVersionKey(desc.CollectionID, desc.VersionID)
		err = txn.Systemstore().Set(ctx, collectionVersionKey.Bytes(), []byte{})
		if err != nil {
			return err
		}
	}

	cache := getCollectionCache(ctx)
	cache.Add(desc)

	return nil
}

func GetCollectionByID(
	ctx context.Context,
	id string,
) (client.CollectionVersion, error) {
	cache := getCollectionCache(ctx)
	col, ok := cache.CollectionsByVersionID[id]
	if ok {
		return col, nil
	}

	txn := datastore.CtxMustGetTxn(ctx)

	key := keys.NewCollectionKey(id)
	buf, err := txn.Systemstore().Get(ctx, key.Bytes())
	if err != nil {
		return client.CollectionVersion{}, err
	}

	err = json.Unmarshal(buf, &col)
	if err != nil {
		return client.CollectionVersion{}, err
	}

	cache.Add(col)

	return col, nil
}

// GetCollectionByName returns the collection with the given name.
//
// If no collection of that name is found, it will return an error.
func GetCollectionByName(
	ctx context.Context,
	name string,
) (client.CollectionVersion, error) {
	cache := getCollectionCache(ctx)
	col, ok := cache.ActiveCollectionsByName[name]
	if ok {
		return col, nil
	}

	txn := datastore.CtxMustGetTxn(ctx)

	nameKey := keys.NewCollectionNameKey(name)
	idBuf, err := txn.Systemstore().Get(ctx, nameKey.Bytes())
	if err != nil {
		return client.CollectionVersion{}, err
	}

	col, err = GetCollectionByID(ctx, string(idBuf))
	if err != nil {
		return client.CollectionVersion{}, err
	}

	cache.Add(col)

	return col, err
}

func GetActiveCollectionByCollectionID(
	ctx context.Context,
	collectionID string,
) (client.CollectionVersion, error) {
	cache := getCollectionCache(ctx)
	col, ok := cache.ActiveCollectionsByID[collectionID]
	if ok {
		return col, nil
	}

	cols, err := GetCollectionsByCollectionID(ctx, collectionID)
	if err != nil {
		return client.CollectionVersion{}, err
	}

	for _, col := range cols {
		if col.IsActive {
			return col, nil
		}
	}

	return client.CollectionVersion{}, corekv.ErrNotFound
}

// GetCollectionsByCollectionID returns all collection versions for the given id.
//
// If no collections are found an empty set will be returned.
func GetCollectionsByCollectionID(
	ctx context.Context,
	collectionID string,
) ([]client.CollectionVersion, error) {
	cache := getCollectionCache(ctx)
	if cache.IsFullyPopulated {
		if col, ok := cache.CollectionsByID[collectionID]; ok {
			return col, nil
		}
		return nil, corekv.ErrNotFound
	}
	// It is not practical to cache a sub set of collections at the moment as figuring
	// out whether the set is complete or not if not possible without fetching the versionIDs
	// anyway.  So we do not cache collections by CollectionID and instead use the cache one-by-one
	// in the GetCollectionByID call.

	versionIDs, err := GetCollectionVersionIDs(ctx, collectionID)
	if err != nil {
		return nil, err
	}

	cols := []client.CollectionVersion{}
	for _, versionID := range versionIDs {
		versionCol, err := GetCollectionByID(ctx, versionID)
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
) ([]client.CollectionVersion, error) {
	cache := getCollectionCache(ctx)
	if cache.IsFullyPopulated {
		return cache.Collections, nil
	}

	txn := datastore.CtxMustGetTxn(ctx)

	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix: []byte(keys.COLLECTION_ID),
	})
	if err != nil {
		return nil, err
	}

	cols := make([]client.CollectionVersion, 0)
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

		var col client.CollectionVersion
		err = json.Unmarshal(value, &col)
		if err != nil {
			if err := iter.Close(); err != nil {
				return nil, NewErrFailedToCloseCollectionQuery(err)
			}
			return nil, err
		}

		cols = append(cols, col)
	}

	cache.AddAll(cols)

	return cols, iter.Close()
}

// GetActiveCollections returns all active collections in the system.
func GetActiveCollections(
	ctx context.Context,
) ([]client.CollectionVersion, error) {
	cache := getCollectionCache(ctx)
	if cache.IsActiveCollectionsPopulated {
		return cache.ActiveCollections, nil
	}

	txn := datastore.CtxMustGetTxn(ctx)

	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix: keys.NewCollectionNameKey("").Bytes(),
	})
	if err != nil {
		return nil, err
	}

	cols := make([]client.CollectionVersion, 0)
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

		col, err := GetCollectionByID(ctx, string(value))
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}

		cols = append(cols, col)
	}

	// Sort the results by ID, so that the order matches that of [GetCollections].
	sort.Slice(cols, func(i, j int) bool { return cols[i].VersionID < cols[j].VersionID })

	cache.AddAllActive(cols)

	return cols, iter.Close()
}

// HasCollectionByName returns true if there is a collection of the given name,
// else returns false.
func HasCollectionByName(
	ctx context.Context,
	name string,
) (bool, error) {
	cache := getCollectionCache(ctx)
	if cache.IsActiveCollectionsPopulated {
		_, ok := cache.ActiveCollectionsByName[name]
		return ok, nil
	}
	txn := datastore.CtxMustGetTxn(ctx)

	nameKey := keys.NewCollectionNameKey(name)
	return txn.Systemstore().Has(ctx, nameKey.Bytes())
}

func GetCollectionVersionIDs(
	ctx context.Context,
	collectionID string,
) ([]string, error) {
	cache := getCollectionCache(ctx)
	if cache.IsFullyPopulated {
		result := []string{}
		if cols, ok := cache.CollectionsByID[collectionID]; ok {
			for _, col := range cols {
				result = append(result, col.VersionID)
			}
			return result, nil
		}
		return nil, corekv.ErrNotFound
	}

	txn := datastore.CtxMustGetTxn(ctx)

	// Add the collection id as the first version here.
	// It is not present in the history prefix.
	collectionIDs := []string{collectionID}

	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix:   keys.NewCollectionVersionKey(collectionID, "").Bytes(),
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

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

		key, err := keys.NewCollectionVersionKeyFromString(string(iter.Key()))
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}

		collectionIDs = append(collectionIDs, key.VersionID)
	}

	return collectionIDs, iter.Close()
}

// GetRelatedCollection returns the collection that the given [FieldKind] points to, if it is found in the
// given [CollectionCache].
//
// If the related collection is not found, default and false will be returned.
func GetRelatedCollection(
	ctx context.Context,
	host client.CollectionVersion,
	kind client.FieldKind,
) (client.CollectionVersion, bool, error) {
	switch typedKind := kind.(type) {
	case *client.NamedKind:
		col, err := GetCollectionByName(ctx, typedKind.Name)
		if errors.Is(err, corekv.ErrNotFound) {
			return client.CollectionVersion{}, false, nil
		}

		return col, true, err

	case *client.CollectionKind:
		col, err := GetActiveCollectionByCollectionID(ctx, typedKind.CollectionID)
		if errors.Is(err, corekv.ErrNotFound) {
			return client.CollectionVersion{}, false, nil
		}

		return col, true, err

	case *client.SelfKind:
		if typedKind.RelativeID == "" {
			return host, true, nil
		}

		cols, err := GetActiveCollections(ctx)
		if err != nil {
			return client.CollectionVersion{}, false, err
		}

		for _, col := range cols {
			if col.CollectionID == host.CollectionID {
				continue
			}

			if col.CollectionSet.Value().CollectionSetID != host.CollectionSet.Value().CollectionSetID {
				continue
			}

			if fmt.Sprint(col.CollectionSet.Value().RelativeID) == typedKind.RelativeID {
				return col, true, nil
			}
		}

	default:
		// no-op
	}

	return client.CollectionVersion{}, false, nil
}

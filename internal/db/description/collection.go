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
	"fmt"
	"sort"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// SaveCollection saves the given collection to the system store.
func SaveCollection(
	ctx context.Context,
	collectionRepository *CollectionRepository,
	desc client.CollectionVersion,
) error {
	return collectionRepository.Write(ctx, desc)
}

func GetCollectionByID(
	ctx context.Context,
	collectionRepository *CollectionRepository,
	id string,
) (client.CollectionVersion, error) {
	col, ok, err := collectionRepository.TryGet(ctx, CollectionIndex{
		Kind:  CollectionVersionID,
		Value: id,
	})
	if err != nil {
		return client.CollectionVersion{}, NewErrGetCollectionByID(err, id)
	}
	if !ok {
		return client.CollectionVersion{}, client.ErrCollectionNotFound
	}
	return col, nil
}

// GetCollectionByName returns the collection with the given name.
//
// If no collection of that name is found, it will return an error.
func GetCollectionByName(
	ctx context.Context,
	collectionRepository *CollectionRepository,
	name string,
) (client.CollectionVersion, error) {
	col, ok, err := collectionRepository.TryGet(ctx, CollectionIndex{
		Kind:  CollectionName,
		Value: name,
	})
	if err != nil {
		return client.CollectionVersion{}, err
	}
	if !ok {
		return client.CollectionVersion{}, client.ErrCollectionNotFound
	}
	return col, nil
}

func GetActiveCollectionByCollectionID(
	ctx context.Context,
	collectionRepository *CollectionRepository,
	colID string,
) (client.CollectionVersion, error) {
	col, ok, err := collectionRepository.TryGet(ctx, CollectionIndex{
		Kind:  CollectionID,
		Value: colID,
	})
	if err != nil {
		return client.CollectionVersion{}, err
	}
	if !ok {
		return client.CollectionVersion{}, client.ErrCollectionNotFound
	}
	return col, nil
}

// GetCollectionsByCollectionID returns all collection versions for the given id.
//
// If no collections are found an empty set will be returned.
func GetCollectionsByCollectionID(
	ctx context.Context,
	collectionRepository *CollectionRepository,
	collectionID string,
) ([]client.CollectionVersion, error) {
	versionIDs, err := getCollectionVersionIDs(ctx, collectionID)
	if err != nil {
		return nil, err
	}

	cols := []client.CollectionVersion{}
	for _, versionID := range versionIDs {
		versionCol, err := GetCollectionByID(ctx, collectionRepository, versionID)
		if err != nil {
			if errors.Is(err, client.ErrCollectionNotFound) {
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
	collectionRepository *CollectionRepository,
) ([]client.CollectionVersion, error) {
	txn := datastore.CtxMustGetTxn(ctx)

	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix:   []byte(keys.COLLECTION_ID),
		KeysOnly: true,
	})
	if err != nil {
		return nil, NewErrGetCollections(err)
	}

	cols := make([]client.CollectionVersion, 0)
	for {
		hasValue, err := iter.Next()
		if err != nil {
			if err := iter.Close(); err != nil {
				return nil, NewErrFailedToCloseCollectionQuery(err)
			}
			return nil, NewErrGetCollections(err)
		}

		if !hasValue {
			break
		}

		key := keys.NewCollectionKeyFromString(string(iter.Key()))

		// We must read via the repository in order to correctly handle collections that may have been
		// deleted by other (committed) transactions - these must not be read from the store.
		col, hasValue, err := collectionRepository.TryGet(
			ctx, CollectionIndex{
				Kind:  CollectionVersionID,
				Value: key.CollectionID,
			},
		)
		if err != nil {
			if errors.Is(err, client.ErrCollectionNotFound) {
				continue
			}
			return nil, NewErrGetCollections(errors.Join(err, iter.Close()))
		}
		if !hasValue {
			continue
		}

		cols = append(cols, col)
	}

	return cols, iter.Close()
}

// GetActiveCollections returns all active collections in the system.
func GetActiveCollections(
	ctx context.Context,
	collectionRepository *CollectionRepository,
) ([]client.CollectionVersion, error) {
	txn := datastore.CtxMustGetTxn(ctx)

	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix: keys.NewCollectionNameKey("").Bytes(),
	})
	if err != nil {
		return nil, NewErrGetActiveCollections(err)
	}

	cols := make([]client.CollectionVersion, 0)
	for {
		hasValue, err := iter.Next()
		if err != nil {
			if err := iter.Close(); err != nil {
				return nil, NewErrFailedToCloseCollectionQuery(err)
			}
			return nil, NewErrGetActiveCollections(err)
		}

		if !hasValue {
			break
		}

		value, err := iter.Value()
		if err != nil {
			if err := iter.Close(); err != nil {
				return nil, NewErrFailedToCloseCollectionQuery(err)
			}
			return nil, NewErrGetActiveCollections(err)
		}

		col, err := GetCollectionByID(ctx, collectionRepository, string(value))
		if err != nil {
			if errors.Is(err, client.ErrCollectionNotFound) {
				continue
			}
			return nil, errors.Join(err, iter.Close())
		}

		cols = append(cols, col)
	}

	// Sort the results by ID, so that the order matches that of [GetCollections].
	sort.Slice(cols, func(i, j int) bool { return cols[i].VersionID < cols[j].VersionID })

	return cols, iter.Close()
}

func getCollectionVersionIDs(
	ctx context.Context,
	collectionID string,
) ([]string, error) {
	txn := datastore.CtxMustGetTxn(ctx)

	// Add the collection id as the first version here.
	// It is not present in the history prefix.
	collectionIDs := []string{collectionID}

	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix:   keys.NewCollectionVersionKey(collectionID, "").Bytes(),
		KeysOnly: true,
	})
	if err != nil {
		return nil, NewErrGetCollectionVersions(err, collectionID)
	}

	for {
		hasValue, err := iter.Next()
		if err != nil {
			if err := iter.Close(); err != nil {
				return nil, NewErrFailedToCloseCollectionQuery(err)
			}
			return nil, NewErrGetCollectionVersions(err, collectionID)
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
	collectionRepository *CollectionRepository,
	host client.CollectionVersion,
	kind client.FieldKind,
) (client.CollectionVersion, bool, error) {
	switch typedKind := kind.(type) {
	case *client.NamedKind:
		col, err := GetCollectionByName(ctx, collectionRepository, typedKind.Name)
		if errors.Is(err, client.ErrCollectionNotFound) {
			return client.CollectionVersion{}, false, nil
		}

		return col, true, err

	case *client.CollectionKind:
		col, err := GetActiveCollectionByCollectionID(ctx, collectionRepository, typedKind.CollectionID)
		if errors.Is(err, client.ErrCollectionNotFound) {
			return client.CollectionVersion{}, false, nil
		}

		return col, true, err

	case *client.SelfKind:
		if typedKind.RelativeID == "" {
			return host, true, nil
		}

		cols, err := GetActiveCollections(ctx, collectionRepository)
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

func DeleteCollection(
	ctx context.Context,
	collectionRepository *CollectionRepository,
	version client.CollectionVersion,
) error {
	return collectionRepository.Delete(ctx, CollectionIndex{
		Kind:  CollectionVersionID,
		Value: version.VersionID,
	})
}

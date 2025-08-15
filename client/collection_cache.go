// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import (
	"context"
	"fmt"

	"github.com/sourcenetwork/immutable"
)

// CollectionCache is an object providing easy access to cached collections.
type CollectionCache struct {
	// The full set of [CollectionVersion]s within this cache
	Collections []CollectionVersion

	// The cached collection versions mapped by their CollectionID
	CollectionsByID map[string]CollectionVersion
}

// NewCollectionCache creates a new [CollectionCache] populated with the given [CollectionVersion]s.
func NewCollectionCache(collections []CollectionVersion) CollectionCache {
	collectionsByID := make(map[string]CollectionVersion, len(collections))

	for _, col := range collections {
		collectionsByID[col.CollectionID] = col
	}

	return CollectionCache{
		Collections:     collections,
		CollectionsByID: collectionsByID,
	}
}

// GetCollection returns the collection that the given [FieldKind] points to, if it is found in the
// given [CollectionCache].
//
// If the related collection is not found, default and false will be returned.
func GetCollection(
	cache CollectionCache,
	host CollectionVersion,
	kind FieldKind,
) (CollectionVersion, bool) {
	switch typedKind := kind.(type) {
	case *NamedKind:
		for _, col := range cache.Collections {
			if col.Name == typedKind.Name {
				return col, true
			}
		}

		return CollectionVersion{}, false

	case *CollectionKind:
		def, ok := cache.CollectionsByID[typedKind.CollectionID]
		return def, ok

	case *SelfKind:
		if typedKind.RelativeID == "" {
			return host, true
		}

		for _, col := range cache.Collections {
			if col.CollectionID == host.CollectionID {
				continue
			}

			if col.CollectionSet.Value().CollectionSetID != host.CollectionSet.Value().CollectionSetID {
				continue
			}

			if fmt.Sprint(col.CollectionSet.Value().RelativeID) == typedKind.RelativeID {
				return col, true
			}
		}

	default:
		// no-op
	}

	return CollectionVersion{}, false
}

// GetCollectionFromStore returns the definition that the given [FieldKind] points to, if it is found
// in the given store.
//
// If the related definition is not found, or an error occurs, default and false will be returned.
func GetCollectionFromStore(
	ctx context.Context,
	store TxnStore,
	host CollectionVersion,
	kind FieldKind,
) (CollectionVersion, bool, error) {
	switch typedKind := kind.(type) {
	case *NamedKind:
		col, err := store.GetCollectionByName(ctx, typedKind.Name)
		if err != nil {
			return CollectionVersion{}, false, err
		}

		return col.Version(), true, nil

	case *CollectionKind:
		cols, err := store.GetCollections(ctx, CollectionFetchOptions{
			CollectionID: immutable.Some(typedKind.CollectionID),
		})

		if len(cols) == 0 {
			return CollectionVersion{}, false, ErrNotFound
		}

		if err != nil {
			return CollectionVersion{}, false, err
		}

		return cols[0].Version(), true, nil

	case *SelfKind:
		if typedKind.RelativeID == "" {
			return host, true, nil
		}

		cols, err := store.GetCollections(ctx, CollectionFetchOptions{
			CollectionSetID: immutable.Some(host.CollectionSet.Value().CollectionSetID),
		})
		if err != nil {
			return CollectionVersion{}, false, err
		}

		for _, col := range cols {
			if col.Version().CollectionID == host.CollectionID {
				continue
			}

			if col.Version().CollectionSet.Value().CollectionSetID != host.CollectionSet.Value().CollectionSetID {
				continue
			}

			if fmt.Sprint(col.Version().CollectionSet.Value().RelativeID) == typedKind.RelativeID {
				return col.Version(), true, nil
			}
		}

	default:
		// no-op
	}

	return CollectionVersion{}, false, nil
}

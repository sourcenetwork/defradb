// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package gen

import (
	"fmt"

	"github.com/sourcenetwork/defradb/client"
)

// CollectionCache is an object providing easy access to cached collections.
type CollectionCache struct {
	// The full set of [CollectionVersion]s within this cache
	Collections []client.CollectionVersion

	// The cached collection versions mapped by their CollectionID
	CollectionsByID map[string]client.CollectionVersion
}

// NewCollectionCache creates a new [CollectionCache] populated with the given [CollectionVersion]s.
func NewCollectionCache(collections []client.CollectionVersion) CollectionCache {
	collectionsByID := make(map[string]client.CollectionVersion, len(collections))

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
	host client.CollectionVersion,
	kind client.FieldKind,
) (client.CollectionVersion, bool) {
	switch typedKind := kind.(type) {
	case *client.NamedKind:
		for _, col := range cache.Collections {
			if col.Name == typedKind.Name {
				return col, true
			}
		}

		return client.CollectionVersion{}, false

	case *client.CollectionKind:
		def, ok := cache.CollectionsByID[typedKind.CollectionID]
		return def, ok

	case *client.SelfKind:
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

	return client.CollectionVersion{}, false
}

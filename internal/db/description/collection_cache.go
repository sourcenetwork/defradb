// Copyright 2025 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/client"
)

type collectionCacheKey struct{}

// InitCollectionCache initialializes the context with a none-nil collection cache.
//
// It is done to avoid an extra check to see if the cache exists or not when fetching
// it from the context.
func InitCollectionCache(ctx context.Context) context.Context {
	return context.WithValue(ctx, collectionCacheKey{}, NewCollectionCache())
}

// getCollectionCache retrieves the collection short-id cache from the given context.
func getCollectionCache(ctx context.Context) *collectionCache {
	return ctx.Value(collectionCacheKey{}).(*collectionCache) //nolint:forcetypeassert
}

// collectionCache is an object providing easy access to cached collections.
type collectionCache struct {
	IsFullyPopulated             bool
	IsActiveCollectionsPopulated bool

	// The cached collection versions mapped by their CollectionID
	ActiveCollectionsByID map[string]client.CollectionVersion

	// The cached collection versions mapped by their CollectionID
	ActiveCollectionsByName map[string]client.CollectionVersion

	// The cached collection versions mapped by their CollectionID
	CollectionsByVersionID map[string]client.CollectionVersion

	// The full set of [CollectionVersion]s within this cache
	Collections       []client.CollectionVersion
	ActiveCollections []client.CollectionVersion
	// The cached collection versions mapped by their CollectionID
	CollectionsByID map[string][]client.CollectionVersion
}

// NewCollectionCache creates a new [collectionCache] populated with the given [CollectionVersion]s.
func NewCollectionCache() *collectionCache {
	return &collectionCache{
		CollectionsByVersionID:  make(map[string]client.CollectionVersion),
		ActiveCollectionsByName: make(map[string]client.CollectionVersion),
		ActiveCollectionsByID:   make(map[string]client.CollectionVersion),
	}
}

func (cache *collectionCache) Add(col client.CollectionVersion) {
	_, isOld := cache.CollectionsByVersionID[col.VersionID]
	cache.CollectionsByVersionID[col.VersionID] = col

	if col.IsActive {
		cache.ActiveCollectionsByName[col.Name] = col
		cache.ActiveCollectionsByID[col.CollectionID] = col
	} else if isOld {
		// If the version already existed in the cache, and the given collection is inactive,
		// ensure that there is no old cached active version
		delete(cache.ActiveCollectionsByID, col.CollectionID)
		delete(cache.ActiveCollectionsByName, col.Name)
	}

	if cache.IsFullyPopulated {
		if !isOld {
			cache.Collections = append(cache.Collections, col)

			colVersions := cache.CollectionsByID[col.CollectionID]
			colVersions = append(colVersions, col)
			cache.CollectionsByID[col.CollectionID] = colVersions
		} else {
			for i, oldC := range cache.Collections {
				if oldC.VersionID == col.VersionID {
					cache.Collections[i] = col
					break
				}
			}

			colVersions := cache.CollectionsByID[col.CollectionID]
			for i := range colVersions {
				if colVersions[i].VersionID == col.VersionID {
					colVersions[i] = col
					break
				}
			}
			cache.CollectionsByID[col.CollectionID] = colVersions
		}
	}

	if cache.IsActiveCollectionsPopulated {
		if !isOld {
			if col.IsActive {
				var found bool
				// If the collection ID already existed in the cache, we need to swap it for the new
				// version
				for i, oldC := range cache.ActiveCollections {
					if oldC.CollectionID == col.CollectionID {
						cache.ActiveCollections[i] = col
						found = true
						break
					}
				}

				if !found {
					cache.ActiveCollections = append(cache.ActiveCollections, col)
				}
			}
		} else {
			if col.IsActive {
				var found bool
				// If the collection version ID already existed in the cache, it may have been patched
				// in which case we need to find and replace the original
				for i, oldC := range cache.ActiveCollections {
					if oldC.VersionID == col.VersionID {
						cache.ActiveCollections[i] = col
						found = true
						break
					}
				}

				if !found {
					cache.ActiveCollections = append(cache.ActiveCollections, col)
				}
			} else {
				for i, oldC := range cache.ActiveCollections {
					if oldC.VersionID == col.VersionID {
						// If the collection has been deactivated, ensure that it is removed from the active set
						cache.ActiveCollections = append(cache.ActiveCollections[:i], cache.ActiveCollections[i+1:]...)
						break
					}
				}
			}
		}
	}
}

func (cache *collectionCache) AddAll(cols []client.CollectionVersion) {
	cache.Collections = make([]client.CollectionVersion, 0, len(cols))
	cache.ActiveCollections = make([]client.CollectionVersion, 0)
	cache.CollectionsByID = make(map[string][]client.CollectionVersion)

	for _, col := range cols {
		cache.Collections = append(cache.Collections, col)
		cache.CollectionsByVersionID[col.VersionID] = col

		colVersions := cache.CollectionsByID[col.CollectionID]
		colVersions = append(colVersions, col)
		cache.CollectionsByID[col.CollectionID] = colVersions

		if col.IsActive {
			cache.ActiveCollectionsByName[col.Name] = col
			cache.ActiveCollectionsByID[col.CollectionID] = col
			cache.ActiveCollections = append(cache.ActiveCollections, col)
		}
	}

	cache.IsFullyPopulated = true
	cache.IsActiveCollectionsPopulated = true
}

func (cache *collectionCache) AddAllActive(cols []client.CollectionVersion) {
	cache.ActiveCollections = make([]client.CollectionVersion, 0, len(cols))

	for _, col := range cols {
		cache.CollectionsByVersionID[col.VersionID] = col

		if col.IsActive {
			cache.ActiveCollectionsByName[col.Name] = col
			cache.ActiveCollectionsByID[col.CollectionID] = col
			cache.ActiveCollections = append(cache.ActiveCollections, col)
		}
	}

	cache.IsActiveCollectionsPopulated = true
}

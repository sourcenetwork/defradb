// Copyright 2026 Democratized Data Foundation
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
	"sync"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/cache"
)

type collectionRepository struct {
	forbiddenLock sync.RWMutex
	// Collections may be forbidden, in which case we must prevent as much interaction with
	// them as possible, regardless of what the underlying corekv store thinks.
	//
	// Collections are forbidden when the last local version is deleted from the node.  They
	// will become unforbidden if they are re-added.
	forbiddenCollectionIDs map[string]struct{}

	// The cached active collection versions mapped by their CollectionID
	activeCollectionsByCollectionID map[string]client.CollectionVersion

	// The cached active collection versions mapped by their Name
	activeCollectionsByName map[string]client.CollectionVersion

	// The cached collection versions mapped by their VersionID.
	//
	// Includes inactive versions.
	collectionsByVersionID map[string]client.CollectionVersion
}

var _ cache.Cache[CollectionIndex, client.CollectionVersion] = (*collectionRepository)(nil)

func newCollectionCache() *collectionRepository {
	return &collectionRepository{
		forbiddenCollectionIDs:          map[string]struct{}{},
		activeCollectionsByCollectionID: map[string]client.CollectionVersion{},
		activeCollectionsByName:         map[string]client.CollectionVersion{},
		collectionsByVersionID:          map[string]client.CollectionVersion{},
	}
}

func (i *collectionRepository) TryGet(key CollectionIndex) (client.CollectionVersion, bool) {
	var col client.CollectionVersion
	var hasValue bool

	switch key.Kind {
	case CollectionVersionID:
		col, hasValue = i.collectionsByVersionID[key.Value]

	case CollectionID:
		col, hasValue = i.activeCollectionsByCollectionID[key.Value]

	case CollectionName:
		col, hasValue = i.activeCollectionsByName[key.Value]
	}

	if hasValue {
		i.forbiddenLock.RLock()
		_, isForbidden := i.forbiddenCollectionIDs[col.CollectionID]
		i.forbiddenLock.RUnlock()

		if isForbidden {
			hasValue = false
			col = client.CollectionVersion{}
		}
	}

	return col, hasValue
}

func (i *collectionRepository) Cache(value client.CollectionVersion) {
	if value.IsActive {
		i.activeCollectionsByName[value.Name] = value
		i.activeCollectionsByCollectionID[value.CollectionID] = value
	} else {
		oldVersion, oldVersionCached := i.collectionsByVersionID[value.VersionID]
		if oldVersionCached && oldVersion.IsActive {
			// If we are deactivating a collection we must remove the old values from the
			// active caches.
			// todo - the unforbidding is currently untested, and should be done as part of:
			// https://github.com/sourcenetwork/defradb/issues/4268
			delete(i.activeCollectionsByName, oldVersion.Name)
			delete(i.activeCollectionsByCollectionID, oldVersion.CollectionID)
		}
	}
	i.collectionsByVersionID[value.VersionID] = value

	i.forbiddenLock.Lock()
	// If this transaction writes a collection version that was previously forbidden, we must unforbid it
	// within the context of this transaction.
	// todo - the unforbidding is currently untested, and should be done as part of:
	// https://github.com/sourcenetwork/defradb/issues/4268
	delete(i.forbiddenCollectionIDs, value.CollectionID)
	i.forbiddenLock.Unlock()
}

func (i *collectionRepository) Remove(key CollectionIndex) {
	value, ok := i.TryGet(key)
	if !ok {
		return
	}

	delete(i.activeCollectionsByName, value.Name)
	delete(i.activeCollectionsByCollectionID, value.CollectionID)
	delete(i.collectionsByVersionID, value.VersionID)
}

func (i *collectionRepository) Forbid(value client.CollectionVersion) {
	i.forbiddenLock.Lock()
	i.forbiddenCollectionIDs[value.CollectionID] = struct{}{}
	i.forbiddenLock.Unlock()
}

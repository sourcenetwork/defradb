// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package keys

const (
	COLLECTION_VIEW_ITEMS = "/collection/vi"
)

// CollectionedKey represents a key that is partially keyed by a collection
// short ID.
//
// This allows access to the key's collection short ID without knowing the underlying
// concrete type.
//
// An important useage of this is to do with acquiring read locks whenever access to
// collection-specific key-values in a store is requested.  Failure to implement this
// interface when declaring Key types can cause parts of the system to bypass held
// collection write locks, potentially resulting in data races and inconsistent action
// results.
type CollectionedKey interface {
	GetCollectionShortID() uint32
}

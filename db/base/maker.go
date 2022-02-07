// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package base

import (
	ds "github.com/ipfs/go-datastore"
	"github.com/sourcenetwork/defradb/core"
)

var (
	ROOT   = "/db"
	SYSTEM = "/db/system"
	DATA   = "/db/data"
	BLOCK  = "/db/block"
	HEAD   = "/db/head"
)

var (
	// Commented because it's deadcode (unused according to linter).
	// collectionSeqKey = "collection"
	collectionNs       = ds.NewKey("/collection/names")
	schemaNs           = ds.NewKey("/schema")
	collectionSchemaNs = ds.NewKey("/collection/schema")
)

// MakeIndexPrefix generates a key prefix for the given collection/index descriptions
func MakeIndexPrefixKey(col *CollectionDescription, index *IndexDescription) core.Key {
	return core.Key{Key: core.NewKey(DATA).ChildString(col.IDString()).ChildString(index.IDString())}
}

// MakeIndexKey generates a key for the target dockey, using the collection/index description
func MakeIndexKey(col *CollectionDescription, index *IndexDescription, key core.Key) core.Key {
	return core.Key{Key: MakeIndexPrefixKey(col, index).Child(key.Key)}
}

// MakeCollectionSystemKey returns a formatted collection key for the system data store.
// it assumes the name of the collection is non-empty.
func MakeCollectionSystemKey(name string) core.Key {
	return core.Key{Key: collectionNs.ChildString(name)}
}

func MakeCollectionSchemaSystemKey(schemaHash string) core.Key {
	return core.Key{Key: collectionSchemaNs.ChildString(schemaHash)}
}

// MakeSchemaSystemKey returns a formatted schema key for the system data store.
// it assumes the name of the schema is non-empty.
func MakeSchemaSystemKey(name string) core.Key {
	return core.Key{Key: schemaNs.ChildString(name)}
}

// MakeIndexPrefixKeyRaw is the same as MakeIndexPrefixKey but it takes as inputs
// the raw datastore keys, instead of the collection and index objects respectively.
func MakeIndexPrefixKeyRaw(collectionID ds.Key, indexID ds.Key) core.Key {
	return core.Key{Key: core.NewKey(DATA).
		Child(collectionID).
		Child(indexID)}
}

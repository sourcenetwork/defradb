package base

import (
	"github.com/sourcenetwork/defradb/core"

	ds "github.com/ipfs/go-datastore"
)

var (
	SYSTEM = "/db/system"
	DATA   = "/db/data"
	BLOCK  = "/db/block"
	HEAD   = "/db/head"
)

var (
	collectionSeqKey = "collection"
	collectionNs     = ds.NewKey("/collection")
)

// MakeIndexPrefix generates a key prefix for the given collection/index descriptions
func MakeIndexPrefixKey(col *CollectionDescription, index *IndexDescription) core.Key {
	return core.Key{core.NewKey(DATA).
		ChildString(col.IDString()).
		ChildString(index.IDString())}
}

// MakeIndexKey generates a key for the target dockey, using the collection/index description
func MakeIndexKey(col *CollectionDescription, index *IndexDescription, key core.Key) core.Key {
	return core.Key{MakeIndexPrefixKey(col, index).Child(key.Key)}
}

// MakeCollectionSystemKey returns a formatted collection key for the system data store.
// it assumes the name of the collection is non-empty.
func MakeCollectionSystemKey(name string) core.Key {
	return core.Key{collectionNs.ChildString(name)}
}

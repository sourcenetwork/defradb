package base

import (
	"github.com/sourcenetwork/defradb/core"
)

var (
	SYSTEM = "/system"
	DATA   = "/data"
	BLOCK  = "/block"
	HEAD   = "/head"
)

// MakeIndexPrefix generates a key prefix for the given collection/index descriptions
func MakeIndexPrefixKey(col *CollectionDescription, index *IndexDescription) core.Key {
	return core.Key{core.NewKey(DATA).
		ChildString(col.IDString()).
		ChildString(index.IDString())}
}

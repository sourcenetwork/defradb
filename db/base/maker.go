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
	"github.com/sourcenetwork/defradb/core"
)

// MakeIndexPrefix generates a key prefix for the given collection/index descriptions
func MakeIndexPrefixKey(col *CollectionDescription, index *IndexDescription) core.DataStoreKey {
	return core.DataStoreKey{
		CollectionId: col.IDString(),
		IndexId:      index.IDString(),
	}
}

// MakeIndexKey generates a key for the target dockey, using the collection/index description
func MakeIndexKey(col *CollectionDescription, index *IndexDescription, key core.DataStoreKey) core.DataStoreKey {
	return core.DataStoreKey{
		CollectionId: col.IDString(),
		IndexId:      index.IDString(),
		DocKey:       key.DocKey,
	}
}

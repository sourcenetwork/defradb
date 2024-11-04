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
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// MakeDataStoreKeyWithCollectionDescription returns the datastore key for the given collection description.
func MakeDataStoreKeyWithCollectionDescription(col client.CollectionDescription) keys.DataStoreKey {
	return keys.DataStoreKey{
		CollectionRootID: col.RootID,
	}
}

// MakeDataStoreKeyWithCollectionAndDocID returns the datastore key for the given docID and collection description.
func MakeDataStoreKeyWithCollectionAndDocID(
	col client.CollectionDescription,
	docID string,
) keys.DataStoreKey {
	return keys.DataStoreKey{
		CollectionRootID: col.RootID,
		DocID:            docID,
	}
}

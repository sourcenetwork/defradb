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
	"errors"
	"fmt"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
)

// MakeIndexPrefix generates a key prefix for the given collection/index descriptions
func MakeIndexPrefixKey(col client.CollectionDescription, index *client.IndexDescription) core.DataStoreKey {
	return core.DataStoreKey{
		CollectionId: col.IDString(),
		IndexId:      index.IDString(),
	}
}

// MakeIndexKey generates a key for the target dockey, using the collection/index description
func MakeIndexKey(col client.CollectionDescription, index *client.IndexDescription, docKey string) core.DataStoreKey {
	return core.DataStoreKey{
		CollectionId: col.IDString(),
		IndexId:      index.IDString(),
		DocKey:       docKey,
	}
}

func MakePrimaryIndexKeyForCRDT(c client.CollectionDescription, ctype client.CType, key core.DataStoreKey, fieldName string) (core.DataStoreKey, error) {
	switch ctype {
	case client.COMPOSITE:
		return MakePrimaryIndexKey(c, key).WithFieldId(core.COMPOSITE_NAMESPACE), nil
	case client.LWW_REGISTER:
		fieldKey := getFieldKey(c, key, fieldName)
		return MakePrimaryIndexKey(c, fieldKey), nil
	}
	return core.DataStoreKey{}, errors.New("Invalid CRDT type")
}

func MakePrimaryIndexKey(c client.CollectionDescription, key core.DataStoreKey) core.DataStoreKey {
	return core.DataStoreKey{
		CollectionId: fmt.Sprint(c.ID),
		IndexId:      fmt.Sprint(c.GetPrimaryIndex().ID),
	}.WithInstanceInfo(key)
}

func getFieldKey(c client.CollectionDescription, key core.DataStoreKey, fieldName string) core.DataStoreKey {
	if !c.IsEmpty() {
		return key.WithFieldId(fmt.Sprint(c.GetFieldKey(fieldName)))
	}
	return key.WithFieldId(fieldName)
}

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
	"fmt"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
)

// MakeIndexPrefix generates a key prefix for the given collection/index descriptions
func MakeCollectionKey(col client.CollectionDescription) core.DataStoreKey {
	return core.DataStoreKey{
		CollectionID: col.IDString(),
	}
}

// MakeIndexKey generates a key for the target dockey, using the collection/index description
func MakeDocKey(col client.CollectionDescription, docKey string) core.DataStoreKey {
	return core.DataStoreKey{
		CollectionID: col.IDString(),
		DocKey:       docKey,
	}
}

func MakePrimaryIndexKeyForCRDT(
	c client.CollectionDescription,
	schema client.SchemaDescription,
	ctype client.CType,
	key core.DataStoreKey,
	fieldName string,
) (core.DataStoreKey, error) {
	switch ctype {
	case client.COMPOSITE:
		return MakeCollectionKey(c).WithInstanceInfo(key).WithFieldId(core.COMPOSITE_NAMESPACE), nil
	case client.LWW_REGISTER:
		field, ok := c.GetFieldByName(fieldName, &schema)
		if !ok {
			return core.DataStoreKey{}, client.NewErrFieldNotExist(fieldName)
		}

		return MakeCollectionKey(c).WithInstanceInfo(key).WithFieldId(fmt.Sprint(field.ID)), nil
	}
	return core.DataStoreKey{}, ErrInvalidCrdtType
}

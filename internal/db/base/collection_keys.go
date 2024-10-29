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
	"github.com/sourcenetwork/defradb/internal/core"
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

func MakePrimaryIndexKeyForCRDT(
	c client.CollectionDefinition,
	ctype client.CType,
	key keys.DataStoreKey,
	fieldName string,
) (keys.DataStoreKey, error) {
	switch ctype {
	case client.COMPOSITE:
		return MakeDataStoreKeyWithCollectionDescription(c.Description).
				WithInstanceInfo(key).
				WithFieldID(core.COMPOSITE_NAMESPACE),
			nil
	case client.LWW_REGISTER, client.PN_COUNTER, client.P_COUNTER:
		field, ok := c.GetFieldByName(fieldName)
		if !ok {
			return keys.DataStoreKey{}, client.NewErrFieldNotExist(fieldName)
		}

		return MakeDataStoreKeyWithCollectionDescription(c.Description).
				WithInstanceInfo(key).
				WithFieldID(fmt.Sprint(field.ID)),
			nil
	}
	return keys.DataStoreKey{}, ErrInvalidCrdtType
}

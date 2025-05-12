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

import (
	"strconv"
	"strings"

	ds "github.com/ipfs/go-datastore"
)

// FieldID indexes field short ids by the full id.
type FieldID struct {
	CollectionShortID uint32
	FieldID           string
}

var _ Key = (*SchemaRootKey)(nil)

func NewFieldID(collectionShortID uint32, fieldID string) FieldID {
	return FieldID{
		CollectionShortID: collectionShortID,
		FieldID:           fieldID,
	}
}

func NewFieldIDP(collectionShortID uint32) FieldID {
	return FieldID{
		CollectionShortID: collectionShortID,
	}
}

func NewFieldIDFromString(keyString string) (FieldID, error) {
	keyString = strings.TrimPrefix(keyString, FIELD_SHORT_ID+"/")
	elements := strings.Split(keyString, "/")
	if len(elements) != 2 {
		return FieldID{}, ErrInvalidKey
	}

	colID, err := strconv.ParseUint(elements[0], 10, 0)
	if err != nil {
		return FieldID{}, err
	}

	return FieldID{
		CollectionShortID: uint32(colID),
		FieldID:           elements[1],
	}, nil
}

func (k FieldID) ToString() string {
	result := FIELD_SHORT_ID

	if k.CollectionShortID != 0 {
		result = result + "/" + strconv.Itoa(int(k.CollectionShortID))
	}

	if k.FieldID != "" {
		result = result + "/" + k.FieldID
	}

	return result
}

func (k FieldID) Bytes() []byte {
	return []byte(k.ToString())
}

func (k FieldID) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

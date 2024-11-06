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
	"fmt"
	"strconv"
	"strings"

	ds "github.com/ipfs/go-datastore"
)

// CollectionSchemaVersionKey points to nil, but the keys/prefix can be used
// to get collections that are using, or have used a given schema version.
//
// If a collection is updated to a different schema version, the old entry(s)
// of this key will be preserved.
//
// This key should be removed in https://github.com/sourcenetwork/defradb/issues/1085
type CollectionSchemaVersionKey struct {
	SchemaVersionID string
	CollectionID    uint32
}

var _ Key = (*CollectionSchemaVersionKey)(nil)

func NewCollectionSchemaVersionKey(schemaVersionId string, collectionID uint32) CollectionSchemaVersionKey {
	return CollectionSchemaVersionKey{
		SchemaVersionID: schemaVersionId,
		CollectionID:    collectionID,
	}
}

func NewCollectionSchemaVersionKeyFromString(key string) (CollectionSchemaVersionKey, error) {
	elements := strings.Split(key, "/")
	colID, err := strconv.Atoi(elements[len(elements)-1])
	if err != nil {
		return CollectionSchemaVersionKey{}, err
	}

	return CollectionSchemaVersionKey{
		SchemaVersionID: elements[len(elements)-2],
		CollectionID:    uint32(colID),
	}, nil
}

func (k CollectionSchemaVersionKey) ToString() string {
	result := COLLECTION_SCHEMA_VERSION

	if k.SchemaVersionID != "" {
		result = result + "/" + k.SchemaVersionID
	}

	if k.CollectionID != 0 {
		result = fmt.Sprintf("%s/%s", result, strconv.Itoa(int(k.CollectionID)))
	}

	return result
}

func (k CollectionSchemaVersionKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k CollectionSchemaVersionKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

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
	"strings"

	ds "github.com/ipfs/go-datastore"
)

// SchemaRootKey indexes schema version ids by their root schema id.
//
// The index is the key, there are no values stored against the key.
type SchemaRootKey struct {
	SchemaRoot      string
	SchemaVersionID string
}

var _ Key = (*SchemaRootKey)(nil)

func NewSchemaRootKey(schemaRoot string, schemaVersionID string) SchemaRootKey {
	return SchemaRootKey{
		SchemaRoot:      schemaRoot,
		SchemaVersionID: schemaVersionID,
	}
}

func NewSchemaRootKeyFromString(keyString string) (SchemaRootKey, error) {
	keyString = strings.TrimPrefix(keyString, SCHEMA_VERSION_ROOT+"/")
	elements := strings.Split(keyString, "/")
	if len(elements) != 2 {
		return SchemaRootKey{}, ErrInvalidKey
	}

	return SchemaRootKey{
		SchemaRoot:      elements[0],
		SchemaVersionID: elements[1],
	}, nil
}

func (k SchemaRootKey) ToString() string {
	result := SCHEMA_VERSION_ROOT

	if k.SchemaRoot != "" {
		result = result + "/" + k.SchemaRoot
	}

	if k.SchemaVersionID != "" {
		result = result + "/" + k.SchemaVersionID
	}

	return result
}

func (k SchemaRootKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k SchemaRootKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

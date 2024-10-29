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

import ds "github.com/ipfs/go-datastore"

// SchemaVersionKey points to the json serialized schema at the specified version.
//
// It's corresponding value is immutable.
type SchemaVersionKey struct {
	SchemaVersionID string
}

var _ Key = (*SchemaVersionKey)(nil)

func NewSchemaVersionKey(schemaVersionID string) SchemaVersionKey {
	return SchemaVersionKey{SchemaVersionID: schemaVersionID}
}

func (k SchemaVersionKey) ToString() string {
	result := SCHEMA_VERSION

	if k.SchemaVersionID != "" {
		result = result + "/" + k.SchemaVersionID
	}

	return result
}

func (k SchemaVersionKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k SchemaVersionKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

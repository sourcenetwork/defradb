// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package schema

import (
	ds "github.com/ipfs/go-datastore"
)

// SchemaStore provides storage for schema, and acts as a cache for generated schemas
//
type SchemaStore struct {
	store ds.Datastore
}

// NewSchemaStore creates a new instance of a schema store
// using the provided Datastore backend.
func NewSchemaStore(backend ds.Datastore) *SchemaStore {
	return nil
}

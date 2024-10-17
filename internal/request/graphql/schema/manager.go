// Copyright 2022 Democratized Data Foundation
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
	gql "github.com/sourcenetwork/graphql-go"
)

// SchemaManager creates an instanced management point
// for schema intake/outtake, and updates.
type SchemaManager struct {
	schema    gql.Schema
	Generator *Generator
}

// NewSchemaManager returns a new instance of a SchemaManager
// with a new default type map
func NewSchemaManager() (*SchemaManager, error) {
	schema, err := defaultSchema()
	if err != nil {
		return nil, err
	}
	sm := &SchemaManager{
		schema: schema,
	}
	sm.NewGenerator()
	return sm, nil
}

func (s *SchemaManager) Schema() *gql.Schema {
	return &s.schema
}

// ResolveTypes ensures all necessary types are defined, and
// resolves any remaining thunks/closures defined on object fields.
// It should be called *after* all dependent types have been added.
func (s *SchemaManager) ResolveTypes() error {
	// basically, this function just refreshes the
	// schema.TypeMap, and runs the internal
	// typeMapReducer (https://github.com/sourcenetwork/graphql-go/blob/v0.7.9/schema.go#L275)
	// which ensures all the necessary types are defined in the
	// typeMap, and if there are any outstanding Thunks/closures
	// resolve them.

	// ATM, there is no function to easily call the internal
	// typeMapReducer function, so as a hack, we are just
	// going to re-add the Query type.

	for _, gqlType := range s.schema.TypeMap() {
		object, isObject := gqlType.(*gql.Object)
		if !isObject {
			continue
		}
		// We need to make sure the object's fields are resolved
		object.Fields()

		if object.Error() != nil {
			return object.Error()
		}
	}

	query := s.schema.QueryType()
	return s.schema.AppendType(query)
}

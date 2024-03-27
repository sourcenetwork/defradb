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

	schemaTypes "github.com/sourcenetwork/defradb/request/graphql/schema/types"
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
	sm := &SchemaManager{}
	schema, err := gql.NewSchema(gql.SchemaConfig{
		Types:      defaultTypes(),
		Query:      defaultQueryType(),
		Mutation:   defaultMutationType(),
		Directives: defaultDirectivesType(),
	})
	if err != nil {
		return sm, err
	}
	sm.schema = schema

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

// @todo: Use a better default Query type
func defaultQueryType() *gql.Object {
	return gql.NewObject(gql.ObjectConfig{
		Name: "Query",
		Fields: gql.Fields{
			"_": &gql.Field{
				Name: "_",
				Type: gql.Boolean,
			},

			// database API queries
			schemaTypes.QueryCommits.Name:       schemaTypes.QueryCommits,
			schemaTypes.QueryLatestCommits.Name: schemaTypes.QueryLatestCommits,
		},
	})
}

func defaultMutationType() *gql.Object {
	return gql.NewObject(gql.ObjectConfig{
		Name: "Mutation",
		Fields: gql.Fields{
			"_": &gql.Field{
				Name: "_",
				Type: gql.Boolean,
			},
		},
	})
}

// default directives type.
func defaultDirectivesType() []*gql.Directive {
	return []*gql.Directive{
		schemaTypes.CRDTFieldDirective,
		schemaTypes.ExplainDirective,
		schemaTypes.PolicyDirective,
		schemaTypes.IndexDirective,
		schemaTypes.IndexFieldDirective,
		schemaTypes.PrimaryDirective,
		schemaTypes.RelationDirective,
	}
}

func inlineArrayTypes() []gql.Type {
	return []gql.Type{
		gql.Boolean,
		gql.Float,
		gql.Int,
		gql.String,
		gql.NewNonNull(gql.Boolean),
		gql.NewNonNull(gql.Float),
		gql.NewNonNull(gql.Int),
		gql.NewNonNull(gql.String),
	}
}

// default type map includes all the native scalar types
func defaultTypes() []gql.Type {
	return []gql.Type{
		// Base Scalar types
		gql.Boolean,
		gql.DateTime,
		gql.Float,
		gql.ID,
		gql.Int,
		gql.String,

		// Custom Scalar types
		schemaTypes.BlobScalarType,
		schemaTypes.JSONScalarType,

		// Base Query types

		// Sort/Order enum
		schemaTypes.OrderingEnum,

		// Filter scalar blocks
		schemaTypes.BooleanOperatorBlock,
		schemaTypes.NotNullBooleanOperatorBlock,
		schemaTypes.DateTimeOperatorBlock,
		schemaTypes.FloatOperatorBlock,
		schemaTypes.NotNullFloatOperatorBlock,
		schemaTypes.IdOperatorBlock,
		schemaTypes.IntOperatorBlock,
		schemaTypes.NotNullIntOperatorBlock,
		schemaTypes.StringOperatorBlock,
		schemaTypes.NotNullstringOperatorBlock,

		schemaTypes.CommitsOrderArg,
		schemaTypes.CommitLinkObject,
		schemaTypes.CommitObject,

		schemaTypes.CRDTEnum,
		schemaTypes.ExplainEnum,
	}
}

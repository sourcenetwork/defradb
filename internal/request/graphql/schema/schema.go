// Copyright 2024 Democratized Data Foundation
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
	"github.com/sourcenetwork/defradb/internal/request/graphql/schema/types"

	gql "github.com/sourcenetwork/graphql-go"
)

// defaultSchema returns a new gql.Schema containing the default type definitions.
func defaultSchema() (gql.Schema, error) {
	orderEnum := types.OrderingEnum()
	commitsEnum := types.CommitsEnum()
	crdtEnum := types.CRDTEnum()
	explainEnum := types.ExplainEnum()

	commitsOrderArg := types.CommitsOrderArg(orderEnum)
	commitsFilterFieldNameArg := types.CommitsFilterFieldNameArg()
	commitsFilterArg := types.CommitsFilterArg(commitsFilterFieldNameArg)

	commitObject := types.CommitObject(commitsOrderArg, commitsFilterArg, commitsEnum)

	encryptedSearchResult := types.EncryptedSearchResultObject()

	indexFieldInput := types.IndexFieldInputObject(orderEnum)

	queryCommits := types.QueryCommits(commitObject, commitsOrderArg, commitsFilterArg, commitsEnum)

	sch, err := gql.NewSchema(gql.SchemaConfig{
		Types: defaultTypes(
			commitObject,
			commitsOrderArg,
			commitsEnum,
			orderEnum,
			crdtEnum,
			explainEnum,
			indexFieldInput,
			encryptedSearchResult,
		),
		Query:        defaultQueryType(queryCommits),
		Mutation:     defaultMutationType(),
		Directives:   defaultDirectivesType(crdtEnum, explainEnum, orderEnum, indexFieldInput),
		Subscription: defaultSubscriptionType(queryCommits),
	})

	return sch, err
}

func defaultQueryType(fields ...*gql.Field) *gql.Object {
	return defaultOperationType("Query", fields...)
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

func defaultSubscriptionType(fields ...*gql.Field) *gql.Object {
	return defaultOperationType("Subscription", fields...)
}

func defaultOperationType(name string, fields ...*gql.Field) *gql.Object {
	fieldsCfg := make(gql.Fields, len(fields))
	for _, field := range fields {
		fieldsCfg[field.Name] = field
	}

	return gql.NewObject(gql.ObjectConfig{
		Name:   name,
		Fields: fieldsCfg,
	})
}

// default directives type.
func defaultDirectivesType(
	crdtEnum *gql.Enum,
	explainEnum *gql.Enum,
	orderEnum *gql.Enum,
	indexFieldInput *gql.InputObject,
) []*gql.Directive {
	return []*gql.Directive{
		types.CRDTFieldDirective(crdtEnum),
		types.DefaultDirective(),
		types.ExhaustiveDirective(),
		types.ExplainDirective(explainEnum),
		types.PolicyDirective(),
		types.IndexDirective(orderEnum, indexFieldInput),
		types.PrimaryDirective(),
		types.RelationDirective(),
		types.MaterializedDirective(),
		types.BranchableDirective(),
		types.VectorEmbeddingDirective(),
		types.ConstraintsDirective(),
		types.EncryptedIndexDirective(),
	}
}

func inlineArrayTypes() []gql.Type {
	return []gql.Type{
		gql.Boolean,
		types.Float32,
		types.Float64,
		gql.Int,
		gql.String,
		gql.NewNonNull(gql.Boolean),
		gql.NewNonNull(gql.Int),
		gql.NewNonNull(gql.String),
		gql.NewNonNull(types.Float32),
		gql.NewNonNull(types.Float64),
	}
}

// default type map includes all the native scalar types
func defaultTypes(
	commitObject *gql.Object,
	commitsOrderArg *gql.InputObject,
	commitsEnum *gql.Enum,
	orderEnum *gql.Enum,
	crdtEnum *gql.Enum,
	explainEnum *gql.Enum,
	indexFieldInput *gql.InputObject,
	encryptedSearchResult *gql.Object,
) []gql.Type {
	idOpBlock := types.IDOperatorBlock()
	intOpBlock := types.IntOperatorBlock()
	float64OpBlock := types.Float64OperatorBlock()
	float32OpBlock := types.Float32OperatorBlock()
	booleanOpBlock := types.BooleanOperatorBlock()
	stringOpBlock := types.StringOperatorBlock()
	blobOpBlock := types.BlobOperatorBlock(types.Blob)
	dateTimeOpBlock := types.DateTimeOperatorBlock()
	scalarAggregateBlock := types.ScalarAggregateNumericBlock()

	notNullIntOpBlock := types.NotNullIntOperatorBlock()
	notNullFloat64OpBlock := types.NotNullFloat64OperatorBlock()
	notNullFloat32OpBlock := types.NotNullFloat32OperatorBlock()
	notNullBooleanOpBlock := types.NotNullBooleanOperatorBlock()
	notNullStringOpBlock := types.NotNullStringOperatorBlock()
	notNullBlobOpBlock := types.NotNullBlobOperatorBlock(types.Blob)

	return []gql.Type{
		// Base Scalar types
		gql.Boolean,
		gql.DateTime,
		gql.Float,
		types.Float32,
		types.Float64,
		gql.ID,
		gql.Int,
		gql.String,

		// Custom Scalar types
		types.Blob,
		types.JSON,

		// Base Query types

		// Sort/Order enum
		orderEnum,

		// Filter scalar blocks
		idOpBlock,
		intOpBlock,
		float64OpBlock,
		float32OpBlock,
		booleanOpBlock,
		stringOpBlock,
		blobOpBlock,
		dateTimeOpBlock,

		// Filter non null scalar blocks
		notNullIntOpBlock,
		notNullFloat64OpBlock,
		notNullFloat32OpBlock,
		notNullBooleanOpBlock,
		notNullStringOpBlock,
		notNullBlobOpBlock,

		// Filter scalar list blocks
		types.IntListOperatorBlock(intOpBlock),
		types.Float64ListOperatorBlock(float64OpBlock),
		types.Float32ListOperatorBlock(float32OpBlock),
		types.BooleanListOperatorBlock(booleanOpBlock),
		types.StringListOperatorBlock(stringOpBlock),

		// Filter non null scalar list blocks
		types.NotNullIntListOperatorBlock(notNullIntOpBlock),
		types.NotNullFloat64ListOperatorBlock(notNullFloat64OpBlock),
		types.NotNullFloat32ListOperatorBlock(notNullFloat32OpBlock),
		types.NotNullBooleanListOperatorBlock(notNullBooleanOpBlock),
		types.NotNullStringListOperatorBlock(notNullStringOpBlock),

		// aggregate input args
		scalarAggregateBlock,

		commitsEnum,
		commitsOrderArg,
		commitObject,

		crdtEnum,
		explainEnum,

		indexFieldInput,
		encryptedSearchResult,
	}
}

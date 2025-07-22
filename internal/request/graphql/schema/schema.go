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
	crdtEnum := types.CRDTEnum()
	explainEnum := types.ExplainEnum()

	commitLinkObject := types.CommitLinkObject()
	commitObject := types.CommitObject(commitLinkObject)
	commitsOrderArg := types.CommitsOrderArg(orderEnum)

	encryptedSearchResult := types.EncryptedSearchResultObject()

	indexFieldInput := types.IndexFieldInputObject(orderEnum)

	return gql.NewSchema(gql.SchemaConfig{
		Types: defaultTypes(
			commitObject,
			commitLinkObject,
			commitsOrderArg,
			orderEnum,
			crdtEnum,
			explainEnum,
			indexFieldInput,
			encryptedSearchResult,
		),
		Query:        defaultQueryType(commitObject, commitsOrderArg),
		Mutation:     defaultMutationType(),
		Directives:   defaultDirectivesType(crdtEnum, explainEnum, orderEnum, indexFieldInput),
		Subscription: defaultSubscriptionType(),
	})
}

// @todo: Use a better default Query type
func defaultQueryType(commitObject *gql.Object, commitsOrderArg *gql.InputObject) *gql.Object {
	queryCommits := types.QueryCommits(commitObject, commitsOrderArg)
	queryLatestCommits := types.QueryLatestCommits(commitObject)

	return gql.NewObject(gql.ObjectConfig{
		Name: "Query",
		Fields: gql.Fields{
			"_": &gql.Field{
				Name: "_",
				Type: gql.Boolean,
			},

			// database API queries
			queryCommits.Name:       queryCommits,
			queryLatestCommits.Name: queryLatestCommits,
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

func defaultSubscriptionType() *gql.Object {
	return gql.NewObject(gql.ObjectConfig{
		Name: "Subscription",
		Fields: gql.Fields{
			"_": &gql.Field{
				Name: "_",
				Type: gql.Boolean,
			},
		},
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
	commitLinkObject *gql.Object,
	commitsOrderArg *gql.InputObject,
	orderEnum *gql.Enum,
	crdtEnum *gql.Enum,
	explainEnum *gql.Enum,
	indexFieldInput *gql.InputObject,
	encryptedSearchResult *gql.Object,
) []gql.Type {
	blobScalarType := types.BlobScalarType()
	jsonScalarType := types.JSONScalarType()

	idOpBlock := types.IDOperatorBlock()
	intOpBlock := types.IntOperatorBlock()
	float64OpBlock := types.Float64OperatorBlock()
	float32OpBlock := types.Float32OperatorBlock()
	booleanOpBlock := types.BooleanOperatorBlock()
	stringOpBlock := types.StringOperatorBlock()
	blobOpBlock := types.BlobOperatorBlock(blobScalarType)
	dateTimeOpBlock := types.DateTimeOperatorBlock()

	notNullIntOpBlock := types.NotNullIntOperatorBlock()
	notNullFloat64OpBlock := types.NotNullFloat64OperatorBlock()
	notNullFloat32OpBlock := types.NotNullFloat32OperatorBlock()
	notNullBooleanOpBlock := types.NotNullBooleanOperatorBlock()
	notNullStringOpBlock := types.NotNullStringOperatorBlock()
	notNullBlobOpBlock := types.NotNullBlobOperatorBlock(blobScalarType)

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
		blobScalarType,
		jsonScalarType,

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

		commitsOrderArg,
		commitLinkObject,
		commitObject,

		crdtEnum,
		explainEnum,

		indexFieldInput,
		encryptedSearchResult,
	}
}

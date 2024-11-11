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
func defaultTypes(
	commitObject *gql.Object,
	commitLinkObject *gql.Object,
	commitsOrderArg *gql.InputObject,
	orderEnum *gql.Enum,
	crdtEnum *gql.Enum,
	explainEnum *gql.Enum,
	indexFieldInput *gql.InputObject,
) []gql.Type {
	blobScalarType := types.BlobScalarType()
	jsonScalarType := types.JSONScalarType()

	idOpBlock := types.IDOperatorBlock()
	intOpBlock := types.IntOperatorBlock()
	floatOpBlock := types.FloatOperatorBlock()
	booleanOpBlock := types.BooleanOperatorBlock()
	stringOpBlock := types.StringOperatorBlock()
	blobOpBlock := types.BlobOperatorBlock(blobScalarType)
	dateTimeOpBlock := types.DateTimeOperatorBlock()

	notNullIntOpBlock := types.NotNullIntOperatorBlock()
	notNullFloatOpBlock := types.NotNullFloatOperatorBlock()
	notNullBooleanOpBlock := types.NotNullBooleanOperatorBlock()
	notNullStringOpBlock := types.NotNullStringOperatorBlock()
	notNullBlobOpBlock := types.NotNullBlobOperatorBlock(blobScalarType)

	return []gql.Type{
		// Base Scalar types
		gql.Boolean,
		gql.DateTime,
		gql.Float,
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
		floatOpBlock,
		booleanOpBlock,
		stringOpBlock,
		blobOpBlock,
		dateTimeOpBlock,

		// Filter non null scalar blocks
		notNullIntOpBlock,
		notNullFloatOpBlock,
		notNullBooleanOpBlock,
		notNullStringOpBlock,
		notNullBlobOpBlock,

		// Filter scalar list blocks
		types.IntListOperatorBlock(intOpBlock),
		types.FloatListOperatorBlock(floatOpBlock),
		types.BooleanListOperatorBlock(booleanOpBlock),
		types.StringListOperatorBlock(stringOpBlock),

		// Filter non null scalar list blocks
		types.NotNullIntListOperatorBlock(notNullIntOpBlock),
		types.NotNullFloatListOperatorBlock(notNullFloatOpBlock),
		types.NotNullBooleanListOperatorBlock(notNullBooleanOpBlock),
		types.NotNullStringListOperatorBlock(notNullStringOpBlock),

		commitsOrderArg,
		commitLinkObject,
		commitObject,

		crdtEnum,
		explainEnum,

		indexFieldInput,
	}
}

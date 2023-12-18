// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package types

import (
	gql "github.com/sourcenetwork/graphql-go"

	"github.com/sourcenetwork/defradb/client/request"
)

var (
	// Helper only for `commit` below.
	commitCountFieldArg = gql.NewEnum(gql.EnumConfig{
		Name:        "commitCountFieldArg",
		Description: CountFieldDescription,
		Values: gql.EnumValueConfigMap{
			"links": &gql.EnumValueConfig{
				Description: commitLinksDescription,
				Value:       "links",
			},
		},
	})

	// Commit represents an individual commit to a MerkleCRDT
	// type Commit {
	// 	Height: Int
	// 	CID: String
	// 	Dockey: String
	// 	CollectionID: Int
	// 	SchemaVersionID: String
	// 	Delta: String
	// 	Previous: [Commit]
	//  Links: [Commit]
	// }
	//
	// Any self referential type needs to be initialized
	// inside the init() func
	CommitObject = gql.NewObject(gql.ObjectConfig{
		Name:        request.CommitTypeName,
		Description: commitDescription,
		Fields: gql.Fields{
			"height": &gql.Field{
				Description: commitHeightFieldDescription,
				Type:        gql.Int,
			},
			"cid": &gql.Field{
				Description: commitCIDFieldDescription,
				Type:        gql.String,
			},
			request.DocID: &gql.Field{
				Description: commitDockeyFieldDescription,
				Type:        gql.String,
			},
			"collectionID": &gql.Field{
				Description: commitCollectionIDFieldDescription,
				Type:        gql.Int,
			},
			"schemaVersionId": &gql.Field{
				Description: commitSchemaVersionIDFieldDescription,
				Type:        gql.String,
			},
			"fieldName": &gql.Field{
				Description: commitFieldNameFieldDescription,
				Type:        gql.String,
			},
			"fieldId": &gql.Field{
				Type:        gql.String,
				Description: commitFieldIDFieldDescription,
			},
			"delta": &gql.Field{
				Description: commitDeltaFieldDescription,
				Type:        gql.String,
			},
			"links": &gql.Field{
				Description: commitLinksDescription,
				Type:        gql.NewList(CommitLinkObject),
			},
			"_count": &gql.Field{
				Description: CountFieldDescription,
				Type:        gql.Int,
				Args: gql.FieldConfigArgument{
					"field": &gql.ArgumentConfig{
						Type: commitCountFieldArg,
					},
				},
			},
		},
	})

	// CommitLink is a named DAG link between commits.
	// This is primary used for CompositeDAG CRDTs
	CommitLinkObject = gql.NewObject(gql.ObjectConfig{
		Name:        "CommitLink",
		Description: commitLinksDescription,
		Fields: gql.Fields{
			"name": &gql.Field{
				Description: commitLinkNameFieldDescription,
				Type:        gql.String,
			},
			"cid": &gql.Field{
				Description: commitLinkCIDFieldDescription,
				Type:        gql.String,
			},
		},
	})

	CommitsOrderArg = gql.NewInputObject(
		gql.InputObjectConfig{
			Name:        "commitsOrderArg",
			Description: OrderArgDescription,
			Fields: gql.InputObjectConfigFieldMap{
				"height": &gql.InputObjectFieldConfig{
					Description: commitHeightFieldDescription,
					Type:        OrderingEnum,
				},
				"cid": &gql.InputObjectFieldConfig{
					Description: commitCIDFieldDescription,
					Type:        OrderingEnum,
				},
				request.DocID: &gql.InputObjectFieldConfig{
					Description: commitDockeyFieldDescription,
					Type:        OrderingEnum,
				},
				"collectionID": &gql.InputObjectFieldConfig{
					Description: commitCollectionIDFieldDescription,
					Type:        OrderingEnum,
				},
			},
		},
	)

	commitFields = gql.NewEnum(
		gql.EnumConfig{
			Name:        "commitFields",
			Description: commitFieldsEnumDescription,
			Values: gql.EnumValueConfigMap{
				"height": &gql.EnumValueConfig{
					Value:       "height",
					Description: commitHeightFieldDescription,
				},
				"cid": &gql.EnumValueConfig{
					Value:       "cid",
					Description: commitCIDFieldDescription,
				},
				request.DocID: &gql.EnumValueConfig{
					Value:       request.DocID,
					Description: commitDockeyFieldDescription,
				},
				"collectionID": &gql.EnumValueConfig{
					Value:       "collectionID",
					Description: commitCollectionIDFieldDescription,
				},
				"fieldName": &gql.EnumValueConfig{
					Value:       "fieldName",
					Description: commitFieldNameFieldDescription,
				},
				"fieldId": &gql.EnumValueConfig{
					Value:       "fieldId",
					Description: commitFieldIDFieldDescription,
				},
			},
		},
	)

	QueryCommits = &gql.Field{
		Name:        "commits",
		Description: commitsQueryDescription,
		Type:        gql.NewList(CommitObject),
		Args: gql.FieldConfigArgument{
			request.DocID:       NewArgConfig(gql.ID, commitDockeyArgDescription),
			request.FieldIDName: NewArgConfig(gql.String, commitFieldIDArgDescription),
			"order":             NewArgConfig(CommitsOrderArg, OrderArgDescription),
			"cid":               NewArgConfig(gql.ID, commitCIDArgDescription),
			"groupBy": NewArgConfig(
				gql.NewList(
					gql.NewNonNull(
						commitFields,
					),
				),
				GroupByArgDescription,
			),
			request.LimitClause:  NewArgConfig(gql.Int, LimitArgDescription),
			request.OffsetClause: NewArgConfig(gql.Int, OffsetArgDescription),
			request.DepthClause:  NewArgConfig(gql.Int, commitDepthArgDescription),
		},
	}

	QueryLatestCommits = &gql.Field{
		Name:        "latestCommits",
		Description: latestCommitsQueryDescription,
		Type:        gql.NewList(CommitObject),
		Args: gql.FieldConfigArgument{
			request.DocID:       NewArgConfig(gql.NewNonNull(gql.ID), commitDockeyArgDescription),
			request.FieldIDName: NewArgConfig(gql.String, commitFieldIDArgDescription),
		},
	}
)

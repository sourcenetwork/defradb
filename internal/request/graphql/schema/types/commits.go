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

// Commit represents an individual commit to a MerkleCRDT
//
//	type Commit {
//		Height: Int
//		CID: String
//		DocID: String
//		CollectionID: Int
//		SchemaVersionID: String
//		Delta: String
//		Previous: [Commit]
//	 Links: [Commit]
//	}
//
// Any self referential type needs to be initialized
// inside the init() func
func CommitObject(commitLinkObject *gql.Object) *gql.Object {
	return gql.NewObject(gql.ObjectConfig{
		Name:        request.CommitTypeName,
		Description: commitDescription,
		Fields: gql.Fields{
			request.HeightFieldName: &gql.Field{
				Description: commitHeightFieldDescription,
				Type:        gql.Int,
			},
			request.CidFieldName: &gql.Field{
				Description: commitCIDFieldDescription,
				Type:        gql.String,
			},
			request.DocIDArgName: &gql.Field{
				Description: commitDocIDFieldDescription,
				Type:        gql.String,
			},
			request.CollectionIDFieldName: &gql.Field{
				Description: commitCollectionIDFieldDescription,
				Type:        gql.Int,
			},
			request.SchemaVersionIDFieldName: &gql.Field{
				Description: commitSchemaVersionIDFieldDescription,
				Type:        gql.String,
			},
			request.FieldNameFieldName: &gql.Field{
				Description: commitFieldNameFieldDescription,
				Type:        gql.String,
			},
			request.FieldIDFieldName: &gql.Field{
				Type:        gql.String,
				Description: commitFieldIDFieldDescription,
			},
			request.DeltaFieldName: &gql.Field{
				Description: commitDeltaFieldDescription,
				Type:        gql.String,
			},
			request.LinksFieldName: &gql.Field{
				Description: commitLinksDescription,
				Type:        gql.NewList(commitLinkObject),
			},
			request.CountFieldName: &gql.Field{
				Description: CountFieldDescription,
				Type:        gql.Int,
				Args: gql.FieldConfigArgument{
					request.FieldName: &gql.ArgumentConfig{
						Type: gql.NewEnum(gql.EnumConfig{
							Name:        "commitCountFieldArg",
							Description: CountFieldDescription,
							Values: gql.EnumValueConfigMap{
								"links": &gql.EnumValueConfig{
									Description: commitLinksDescription,
									Value:       "links",
								},
							},
						}),
					},
				},
			},
		},
	})
}

// CommitLink is a named DAG link between commits.
// This is primary used for CompositeDAG CRDTs
func CommitLinkObject() *gql.Object {
	return gql.NewObject(gql.ObjectConfig{
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
}

func CommitsOrderArg(orderEnum *gql.Enum) *gql.InputObject {
	return gql.NewInputObject(
		gql.InputObjectConfig{
			Name:        "commitsOrderArg",
			Description: OrderArgDescription,
			Fields: gql.InputObjectConfigFieldMap{
				"height": &gql.InputObjectFieldConfig{
					Description: commitHeightFieldDescription,
					Type:        orderEnum,
				},
				"cid": &gql.InputObjectFieldConfig{
					Description: commitCIDFieldDescription,
					Type:        orderEnum,
				},
				request.DocIDArgName: &gql.InputObjectFieldConfig{
					Description: commitDocIDFieldDescription,
					Type:        orderEnum,
				},
				"collectionID": &gql.InputObjectFieldConfig{
					Description: commitCollectionIDFieldDescription,
					Type:        orderEnum,
				},
			},
		},
	)
}

func QueryCommits(commitObject *gql.Object, commitsOrderArg *gql.InputObject) *gql.Field {
	return &gql.Field{
		Name:        "commits",
		Description: commitsQueryDescription,
		Type:        gql.NewList(commitObject),
		Args: gql.FieldConfigArgument{
			request.DocIDArgName: NewArgConfig(gql.ID, commitDocIDArgDescription),
			request.FieldIDName:  NewArgConfig(gql.String, commitFieldIDArgDescription),
			"order":              NewArgConfig(gql.NewList(commitsOrderArg), OrderArgDescription),
			"cid":                NewArgConfig(gql.ID, commitCIDArgDescription),
			"groupBy": NewArgConfig(
				gql.NewList(
					gql.NewNonNull(
						gql.NewEnum(
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
									request.DocIDArgName: &gql.EnumValueConfig{
										Value:       request.DocIDArgName,
										Description: commitDocIDFieldDescription,
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
						),
					),
				),
				GroupByArgDescription,
			),
			request.LimitClause:  NewArgConfig(gql.Int, LimitArgDescription),
			request.OffsetClause: NewArgConfig(gql.Int, OffsetArgDescription),
			request.DepthClause:  NewArgConfig(gql.Int, commitDepthArgDescription),
		},
	}
}

func QueryLatestCommits(commitObject *gql.Object) *gql.Field {
	return &gql.Field{
		Name:        "latestCommits",
		Description: latestCommitsQueryDescription,
		Type:        gql.NewList(commitObject),
		Args: gql.FieldConfigArgument{
			request.DocIDArgName: NewArgConfig(gql.NewNonNull(gql.ID), commitDocIDArgDescription),
			request.FieldIDName:  NewArgConfig(gql.String, commitFieldIDArgDescription),
		},
	}
}

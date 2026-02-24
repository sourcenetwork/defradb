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
	"github.com/sourcenetwork/defradb/internal/connor"
)

// Commit represents an individual commit to a MerkleCRDT
//
//	type Commit {
//		Height: Int
//		CID: String
//		DocID: String
//		CollectionID: Int
//		CollectionVersionID: String
//		Delta: String
//		Links: [Commit]
//		Heads: [Commit]
//		Signature: Signature
//	}
//
//	type Signature {
//		Type: String
//		Identity: String
//		Value: String
//	}
//
// Any self referential type needs to be initialized
// inside the init() func
func CommitObject(
	commitsOrderArg *gql.InputObject,
	commitsFilterArg *gql.InputObject,
	commitsEnum *gql.Enum,
) *gql.Object {
	// we need the fieldThunk since we are creating a circular type reference
	var commitObject *gql.Object
	fieldsThunk := (gql.FieldsThunk)(func() (gql.Fields, error) {
		commitLinkType := &gql.Field{
			Description: commitLinksDescription,
			Type:        gql.NewList(commitObject),
			Args: gql.FieldConfigArgument{
				request.DocIDArgName: NewArgConfig(gql.ID, commitDocIDArgDescription),
				request.FilterClause: NewArgConfig(commitsFilterArg, "Filter results based on specified conditions."),
				"order":              NewArgConfig(gql.NewList(commitsOrderArg), OrderArgDescription),
				request.CidArgName:   NewArgConfig(gql.NewList(gql.NewNonNull(gql.ID)), commitCIDArgDescription),
				"groupBy": NewArgConfig(
					gql.NewList(
						gql.NewNonNull(
							commitsEnum,
						),
					),
					GroupByArgDescription,
				),
			},
		}

		fields := gql.Fields{
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
			request.CollectionVersionIDFieldName: &gql.Field{
				Description: commitCollectionVersionIDFieldDescription,
				Type:        gql.String,
			},
			request.FieldNameName: &gql.Field{
				Description: commitFieldNameFieldDescription,
				Type:        gql.String,
			},
			request.DeltaFieldName: &gql.Field{
				Description: commitDeltaFieldDescription,
				Type:        gql.String,
			},
			request.LinksFieldName: commitLinkType,
			request.HeadsFieldName: commitLinkType,
			request.SignatureFieldName: &gql.Field{
				Description: signatureDescription,
				Type: gql.NewObject(gql.ObjectConfig{
					Name:        request.SignatureTypeName,
					Description: signatureDescription,
					Fields: gql.Fields{
						request.SignatureTypeFieldName: &gql.Field{
							Description: "The type of the signature, which is used to determine the " +
								"algorithm used to generate the signature.",
							Type: gql.String,
						},
						request.SignatureIdentityFieldName: &gql.Field{
							Description: "The identity of the signer, which is used to determine " +
								"the public key used to verify the signature.",
							Type: gql.String,
						},
						request.SignatureValueFieldName: &gql.Field{
							Description: "The value of the signature, which is used to verify the integrity " +
								"of the commit and the data it contains.",
							Type: gql.String,
						},
					}},
				),
			},
			request.CountFieldName: &gql.Field{
				Description: CountFieldDescription,
				Type:        gql.Int,
				Args: gql.FieldConfigArgument{
					request.FieldArgName: &gql.ArgumentConfig{
						Type: gql.NewEnum(gql.EnumConfig{
							Name:        "commitCountFieldArg",
							Description: CountFieldDescription,
							Values: gql.EnumValueConfigMap{
								"links": &gql.EnumValueConfig{
									Description: commitLinksDescription,
									Value:       "links",
								},
								"heads": &gql.EnumValueConfig{
									Description: commitLinksDescription,
									Value:       "heads",
								},
							},
						}),
					},
				},
			},
		}
		return fields, nil
	})

	commitObject = gql.NewObject(gql.ObjectConfig{
		Name:        request.CommitTypeName,
		Description: commitDescription,
		Fields:      fieldsThunk,
	})
	return commitObject
}

func CommitsFilterFieldNameArg() *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "CommitsFieldNameFilterArg",
		Description: "Filter operators for commit fieldName.",
		Fields: gql.InputObjectConfigFieldMap{
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        gql.String,
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        gql.String,
			},
			"_in": &gql.InputObjectFieldConfig{
				Description: inOperatorDescription,
				Type:        gql.NewList(gql.String),
			},
			"_nin": &gql.InputObjectFieldConfig{
				Description: ninOperatorDescription,
				Type:        gql.NewList(gql.String),
			},
		},
	})
}

func CommitsFilterArg(fieldNameFilter *gql.InputObject) *gql.InputObject {
	var selfRefType *gql.InputObject

	inputCfg := gql.InputObjectConfig{
		Name:        "CommitsFilterArg",
		Description: "Filter argument for commits query.",
	}

	fieldThunk := (gql.InputObjectConfigFieldMapThunk)(
		func() (gql.InputObjectConfigFieldMap, error) {
			return gql.InputObjectConfigFieldMap{
				request.FieldNameName: &gql.InputObjectFieldConfig{
					Description: "Filter commits by field name. Use \"_C\" for document composite commits, " +
						"the field name (e.g. \"age\") for field commits, " +
						"or null for collection commits on branchable collections.",
					Type: fieldNameFilter,
				},
				request.FilterOpAnd: &gql.InputObjectFieldConfig{
					Description: AndOperatorDescription,
					Type:        gql.NewList(gql.NewNonNull(selfRefType)),
				},
				request.FilterOpOr: &gql.InputObjectFieldConfig{
					Description: OrOperatorDescription,
					Type:        gql.NewList(gql.NewNonNull(selfRefType)),
				},
			}, nil
		},
	)

	inputCfg.Fields = fieldThunk
	selfRefType = gql.NewInputObject(inputCfg)
	return selfRefType
}

func CommitsEnum() *gql.Enum {
	return gql.NewEnum(
		gql.EnumConfig{
			Name:        "commitFields",
			Description: commitFieldsEnumDescription,
			Values: gql.EnumValueConfigMap{
				request.HeightArgName: &gql.EnumValueConfig{
					Value:       request.HeightArgName,
					Description: commitHeightFieldDescription,
				},
				request.CidArgName: &gql.EnumValueConfig{
					Value:       request.CidArgName,
					Description: commitCIDFieldDescription,
				},
				request.DocIDArgName: &gql.EnumValueConfig{
					Value:       request.DocIDArgName,
					Description: commitDocIDFieldDescription,
				},
				request.FieldNameName: &gql.EnumValueConfig{
					Value:       request.FieldNameName,
					Description: commitFieldNameFieldDescription,
				},
			},
		},
	)
}

func CommitsOrderArg(orderEnum *gql.Enum) *gql.InputObject {
	return gql.NewInputObject(
		gql.InputObjectConfig{
			Name:        "commitsOrderArg",
			Description: OrderArgDescription,
			Fields: gql.InputObjectConfigFieldMap{
				request.HeightArgName: &gql.InputObjectFieldConfig{
					Description: commitHeightFieldDescription,
					Type:        orderEnum,
				},
				request.CidArgName: &gql.InputObjectFieldConfig{
					Description: commitCIDFieldDescription,
					Type:        orderEnum,
				},
				request.DocIDArgName: &gql.InputObjectFieldConfig{
					Description: commitDocIDFieldDescription,
					Type:        orderEnum,
				},
			},
		},
	)
}

func QueryCommits(
	commitObject *gql.Object,
	commitsOrderArg *gql.InputObject,
	commitsFilterArg *gql.InputObject,
	commitsEnum *gql.Enum,
) *gql.Field {
	return &gql.Field{
		Name:        request.CommitsName,
		Description: commitsQueryDescription,
		Type:        gql.NewList(commitObject),
		Args: gql.FieldConfigArgument{
			request.DocIDArgName: NewArgConfig(gql.ID, commitDocIDArgDescription),
			request.FilterClause: NewArgConfig(commitsFilterArg, "Filter results based on specified conditions."),
			"order":              NewArgConfig(gql.NewList(commitsOrderArg), OrderArgDescription),
			request.CidArgName:   NewArgConfig(gql.NewList(gql.NewNonNull(gql.ID)), commitCIDArgDescription),
			"groupBy": NewArgConfig(
				gql.NewList(
					gql.NewNonNull(
						commitsEnum,
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

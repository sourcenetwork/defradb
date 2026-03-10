// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package collection_version

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestInputTypeOfOrderFieldWhereCollectionHasManyRelationType(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type user {
						age: Int
						name: String
						points: Float
						verified: Boolean
						group: group
					}

					type group {
						members: [user]
					}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "group") {
							name
							fields {
								name
								args {
									name
									type {
										name
										ofType {
											name
											kind
										}
										inputFields {
											name
											type {
												name
												ofType {
													name
													kind
												}
											}
										}
									}
								}
							}
						}
					}
				`,
				ContainsData: map[string]any{
					"__type": map[string]any{
						"name": "group",
						"fields": []any{
							map[string]any{
								// Asserting only on group, because it is the field that contains `order` info we are
								// looking for, additionally wanted to reduce the noise of other elements that were getting
								// dumped out which made the entire output horrible.
								"name": "GROUP",
								"args": append(
									trimFields(
										fields{
											docIDArg,
											buildFilterArg("group", []argDef{
												{
													fieldName: "members",
													typeName:  "userFilterArg",
												},
											}),
											groupByArg,
											limitArg,
											offsetArg,
										},
										testInputTypeOfOrderFieldWhereCollectionHasRelationTypeArgProps,
									),
									map[string]any{
										"name": "order",
										"type": map[string]any{
											"name": nil,
											"ofType": map[string]any{
												"kind": "INPUT_OBJECT",
												"name": "groupOrderArg",
											},
											"inputFields": nil,
										},
									},
								).Tidy(),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestInputTypeOfOrderFieldWhereCollectionHasRelationType(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type book {
						name: String
						rating: Float
						author: author
					}

					type author {
						name: String
						age: Int
						verified: Boolean
						wrote: book @primary
					}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "author") {
							name
							fields {
								name
								args {
									name
									type {
										name
										ofType {
											name
											kind
										}
										inputFields {
											name
											type {
												name
												ofType {
													name
													kind
												}
											}
										}
									}
								}
							}
						}
					}
				`,
				ContainsData: map[string]any{
					"__type": map[string]any{
						"name": "author",
						"fields": []any{
							map[string]any{
								// Asserting only on group, because it is the field that contains `order` info we are
								// looking for, additionally wanted to reduce the noise of other elements that were getting
								// dumped out which made the entire output horrible.
								"name": "GROUP",
								"args": append(
									defaultGroupArgsWithoutOrder,
									map[string]any{
										"name": "order",
										"type": map[string]any{
											"name":        nil,
											"inputFields": nil,
											"ofType": map[string]any{
												"kind": "INPUT_OBJECT",
												"name": "authorOrderArg",
											},
										},
									},
								).Tidy(),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

var testInputTypeOfOrderFieldWhereCollectionHasRelationTypeArgProps = map[string]any{
	"name": struct{}{},
	"type": map[string]any{
		"name": struct{}{},
		"ofType": map[string]any{
			"kind": struct{}{},
			"name": struct{}{},
		},
		"inputFields": struct{}{},
	},
}

var defaultGroupArgsWithoutOrder = trimFields(
	fields{
		docIDArg,
		buildFilterArg("author", []argDef{
			{
				fieldName: "_wroteID",
				typeName:  "IDOperatorBlock",
			},
			{
				fieldName: "age",
				typeName:  "IntOperatorBlock",
			},
			{
				fieldName: "name",
				typeName:  "StringOperatorBlock",
			},
			{
				fieldName: "verified",
				typeName:  "BooleanOperatorBlock",
			},
			{
				fieldName: "wrote",
				typeName:  "bookFilterArg",
			},
		}),
		groupByArg,
		limitArg,
		offsetArg,
	},
	testInputTypeOfOrderFieldWhereCollectionHasRelationTypeArgProps,
)

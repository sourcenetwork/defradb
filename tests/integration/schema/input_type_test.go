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
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestInputTypeOfOrderFieldWhereSchemaHasManyRelationType(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
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
								"name": "_group",
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
										testInputTypeOfOrderFieldWhereSchemaHasRelationTypeArgProps,
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

func TestInputTypeOfOrderFieldWhereSchemaHasRelationType(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
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
								"name": "_group",
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

var testInputTypeOfOrderFieldWhereSchemaHasRelationTypeArgProps = map[string]any{
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
			{
				fieldName: "wrote_id",
				typeName:  "IDOperatorBlock",
			},
		}),
		groupByArg,
		limitArg,
		offsetArg,
	},
	testInputTypeOfOrderFieldWhereSchemaHasRelationTypeArgProps,
)

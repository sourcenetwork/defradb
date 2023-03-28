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

/* WIP "AuthorFilterArg" vs "authorFilterArg".
func TestInputTypeOfOrderFieldWhereSchemaHasRelationType(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Book {
						name: String
						rating: Float
						author: Author
					}

					type Author {
						name: String
						age: Int
						verified: Boolean
						wrote: Book @primary
					}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "Author") {
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
						"name": "Author",
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
											"name":   "AuthorOrderArg",
											"ofType": nil,
											"inputFields": []any{
												map[string]any{
													"name": "_key",
													"type": map[string]any{
														"name":   "Ordering",
														"ofType": nil,
													},
												},
												map[string]any{
													"name": "age",
													"type": map[string]any{
														"name":   "Ordering",
														"ofType": nil,
													},
												},
												map[string]any{
													"name": "name",
													"type": map[string]any{
														"name":   "Ordering",
														"ofType": nil,
													},
												},
												map[string]any{
													"name": "verified",
													"type": map[string]any{
														"name":   "Ordering",
														"ofType": nil,
													},
												},
												// Without the relation type we won't have the following ordering type(s).
												map[string]any{
													"name": "wrote",
													"type": map[string]any{
														"name":   "BookOrderArg",
														"ofType": nil,
													},
												},
												map[string]any{
													"name": "wrote_id",
													"type": map[string]any{
														"name":   "Ordering",
														"ofType": nil,
													},
												},
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

	testUtils.ExecuteTestCase(t, []string{"Book", "Author"}, test)
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
		dockeyArg,
		dockeysArg,
		buildFilterArg("Author", []argDef{
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
				typeName:  "BookFilterArg",
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
*/

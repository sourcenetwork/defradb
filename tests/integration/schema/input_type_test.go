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
)

func TestInputTypeOfComplexSchema(t *testing.T) {
	test := QueryTestCase{
		Schema: []string{
			`
				type book {
				    name: String
				    rating: Float
				    author: author
				    publisher: publisher
				}

				type author {
				    name: String
				    age: Int
				    verified: Boolean
				    wrote: book @primary
				}
				
				type publisher {
				    name: String
				    address: String
				    favouritePageNumbers: [Int!]
				    published: [book]
				}
			`,
		},
		IntrospectionQuery: `
			query IntrospectionQuery {
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
		ContainsData: map[string]interface{}{
			"__type": map[string]interface{}{
				"name": "author",
				"fields": []interface{}{
					map[string]interface{}{
						"name": "_group",
						"args": []interface{}{
							map[string]interface{}{
								"name": "filter",
								"type": map[string]interface{}{
									"name":   "authorFilterArg",
									"ofType": interface{}(nil),
									"inputFields": []interface{}{
										map[string]interface{}{
											"name": "_and",
											"type": map[string]interface{}{
												"name": interface{}(nil),
												"ofType": map[string]interface{}{
													"kind": "INPUT_OBJECT",
													"name": "authorFilterArg",
												},
											},
										},
										map[string]interface{}{
											"name": "_key",
											"type": map[string]interface{}{
												"name":   "IDOperatorBlock",
												"ofType": interface{}(nil),
											},
										},
										map[string]interface{}{
											"name": "_not",
											"type": map[string]interface{}{
												"name":   "authorFilterArg",
												"ofType": interface{}(nil)},
										},
										map[string]interface{}{
											"name": "_or",
											"type": map[string]interface{}{
												"name": interface{}(nil),
												"ofType": map[string]interface{}{
													"kind": "INPUT_OBJECT",
													"name": "authorFilterArg",
												},
											},
										},
										map[string]interface{}{
											"name": "age",
											"type": map[string]interface{}{
												"name":   "IntOperatorBlock",
												"ofType": interface{}(nil),
											},
										},
										map[string]interface{}{
											"name": "name",
											"type": map[string]interface{}{
												"name":   "StringOperatorBlock",
												"ofType": interface{}(nil),
											},
										},
										map[string]interface{}{
											"name": "verified",
											"type": map[string]interface{}{
												"name":   "BooleanOperatorBlock",
												"ofType": interface{}(nil),
											},
										},
										map[string]interface{}{
											"name": "wrote",
											"type": map[string]interface{}{
												"name":   "bookFilterBaseArg",
												"ofType": interface{}(nil),
											},
										},
										map[string]interface{}{
											"name": "wrote_id",
											"type": map[string]interface{}{
												"name":   "IDOperatorBlock",
												"ofType": interface{}(nil),
											},
										},
									},
								},
							},
							map[string]interface{}{
								"name": "groupBy",
								"type": map[string]interface{}{
									"inputFields": interface{}(nil),
									"name":        interface{}(nil),
									"ofType": map[string]interface{}{
										"kind": "NON_NULL",
										"name": interface{}(nil),
									},
								},
							},
							map[string]interface{}{
								"name": "having",
								"type": map[string]interface{}{
									"name":   "authorHavingArg",
									"ofType": interface{}(nil),
									"inputFields": []interface{}{
										map[string]interface{}{
											"name": "_avg",
											"type": map[string]interface{}{
												"name":   "authorHavingBlock",
												"ofType": interface{}(nil),
											},
										},
										map[string]interface{}{
											"name": "_count",
											"type": map[string]interface{}{
												"name":   "authorHavingBlock",
												"ofType": interface{}(nil),
											},
										},
										map[string]interface{}{
											"name": "_key",
											"type": map[string]interface{}{
												"name":   "authorHavingBlock",
												"ofType": interface{}(nil),
											},
										},
										map[string]interface{}{
											"name": "_sum",
											"type": map[string]interface{}{
												"name":   "authorHavingBlock",
												"ofType": interface{}(nil),
											},
										},
										map[string]interface{}{
											"name": "age",
											"type": map[string]interface{}{
												"name":   "authorHavingBlock",
												"ofType": interface{}(nil),
											},
										},
										map[string]interface{}{
											"name": "name",
											"type": map[string]interface{}{
												"name":   "authorHavingBlock",
												"ofType": interface{}(nil),
											},
										},
										map[string]interface{}{
											"name": "verified",
											"type": map[string]interface{}{
												"name":   "authorHavingBlock",
												"ofType": interface{}(nil),
											},
										},
										map[string]interface{}{
											"name": "wrote_id",
											"type": map[string]interface{}{
												"name":   "authorHavingBlock",
												"ofType": interface{}(nil),
											},
										},
									},
								},
							},
							map[string]interface{}{
								"name": "limit",
								"type": map[string]interface{}{
									"inputFields": interface{}(nil),
									"name":        "Int",
									"ofType":      interface{}(nil),
								},
							},
							map[string]interface{}{
								"name": "offset",
								"type": map[string]interface{}{
									"inputFields": interface{}(nil),
									"name":        "Int",
									"ofType":      interface{}(nil),
								},
							},
							map[string]interface{}{
								"name": "order",
								"type": map[string]interface{}{
									"inputFields": []interface{}{
										map[string]interface{}{
											"name": "_key",
											"type": map[string]interface{}{
												"name":   "Ordering",
												"ofType": interface{}(nil),
											},
										},
										map[string]interface{}{
											"name": "age",
											"type": map[string]interface{}{
												"name":   "Ordering",
												"ofType": interface{}(nil),
											},
										},
										map[string]interface{}{
											"name": "name",
											"type": map[string]interface{}{
												"name":   "Ordering",
												"ofType": interface{}(nil),
											},
										},
										map[string]interface{}{
											"name": "verified",
											"type": map[string]interface{}{
												"name":   "Ordering",
												"ofType": interface{}(nil),
											},
										},
										map[string]interface{}{
											"name": "wrote",
											"type": map[string]interface{}{
												"name":   "bookOrderArg",
												"ofType": interface{}(nil),
											},
										},
										map[string]interface{}{
											"name": "wrote_id",
											"type": map[string]interface{}{
												"name":   "Ordering",
												"ofType": interface{}(nil),
											},
										},
									},
									"name":   "authorOrderArg",
									"ofType": interface{}(nil),
								},
							},
						},
					},
				},
			},
		},
	}

	ExecuteQueryTestCase(t, test)
}

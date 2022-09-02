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

func TestInputTypeOfOrderFieldWhereSchemaHasRelationType(t *testing.T) {
	test := QueryTestCase{
		Schema: []string{
			`
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
									"name":   "authorOrderArg",
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
												"name":   "bookOrderArg",
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
						),
					},
				},
			},
		},
	}

	ExecuteQueryTestCase(t, test)
}

var defaultGroupArgsWithoutOrder = []any{

	map[string]any{
		"name": "filter",
		"type": filterArg,
	},

	map[string]any{
		"name": "groupBy",
		"type": groupByArg,
	},

	map[string]any{
		"name": "having",
		"type": havingArg,
	},

	map[string]any{
		"name": "limit",
		"type": intArgType,
	},

	map[string]any{
		"name": "offset",
		"type": intArgType,
	},
}

var groupByArg = map[string]any{
	"inputFields": nil,
	"name":        nil,
	"ofType": map[string]any{
		"kind": "NON_NULL",
		"name": nil,
	},
}

var intArgType = map[string]any{
	"inputFields": nil,
	"name":        "Int",
	"ofType":      nil,
}

var filterArg = map[string]any{
	"name":   "authorFilterArg",
	"ofType": nil,
	"inputFields": []any{
		makeInputObject("_and", nil, inputObjAuthorFilterArg),
		makeInputObject("_key", "IDOperatorBlock", nil),
		makeInputObject("_not", "authorFilterArg", nil),
		makeInputObject("_or", nil, inputObjAuthorFilterArg),
		makeInputObject("age", "IntOperatorBlock", nil),
		makeInputObject("name", "StringOperatorBlock", nil),
		makeInputObject("verified", "BooleanOperatorBlock", nil),
		makeInputObject("wrote", "bookFilterBaseArg", nil),
		makeInputObject("wrote_id", "IDOperatorBlock", nil),
	},
}

var havingArg = map[string]any{
	"name":   "authorHavingArg",
	"ofType": nil,
	"inputFields": []any{
		makeAuthorHavingBlockForName("_avg"),
		makeAuthorHavingBlockForName("_count"),
		makeAuthorHavingBlockForName("_key"),
		makeAuthorHavingBlockForName("_sum"),
		makeAuthorHavingBlockForName("age"),
		makeAuthorHavingBlockForName("name"),
		makeAuthorHavingBlockForName("verified"),
		makeAuthorHavingBlockForName("wrote_id"),
	},
}

var inputObjAuthorFilterArg = map[string]any{
	"kind": "INPUT_OBJECT",
	"name": "authorFilterArg",
}

func makeAuthorHavingBlockForName(name string) map[string]any {
	return makeInputObject(name, "authorHavingBlock", nil)
}

// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"

	"github.com/sourcenetwork/immutable"
)

func TestQuerySimpleWithNonNullVariable(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with non null variable",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21
				}`,
			},
			testUtils.Request{
				Variables: immutable.Some(map[string]any{
					"age": 50,
					"ord": "ASC",
				}),
				Request: `query($age: Int!, $ord: Ordering!) {
					Users(filter: {Age: {_lt: $age}}, order: {Age: $ord}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
						},
						{
							"Name": "Alice",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithVariableDefaultValue(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with variable default value",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21
				}`,
			},
			testUtils.Request{
				Request: `query($age: Int = 50, $ord: Ordering = ASC) {
					Users(filter: {Age: {_lt: $age}}, order: {Age: $ord}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
						},
						{
							"Name": "Alice",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithNonNullVariable_ReturnsErrorWhenNull(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with non null variable returns error when null",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21
				}`,
			},
			testUtils.Request{
				Request: `query($age: Int!) {
					Users(filter: {Age: {_lt: $age}}) {
						Name
					}
				}`,
				ExpectedError: "Variable \"$age\" of required type \"Int!\" was not provided.",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithVariableDefaultValueOverride(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with variable default value override",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21
				}`,
			},
			testUtils.Request{
				Variables: immutable.Some(map[string]any{
					"age": int64(30),
				}),
				Request: `query($age: Int = 50) {
					Users(filter: {Age: {_lt: $age}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithOrderVariable(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with order variable",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21
				}`,
			},
			testUtils.Request{
				Variables: immutable.Some(map[string]any{
					"order": []map[string]any{
						{"Name": "DESC"},
						{"Age": "ASC"},
					},
				}),
				Request: `query($order: [UsersOrderArg]) {
					Users(order: $order) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
						},
						{
							"Name": "Alice",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithAggregateCountVariable(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with aggregate count variable",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21
				}`,
			},
			testUtils.Request{
				Variables: immutable.Some(map[string]any{
					"usersCount": map[string]any{
						"filter": map[string]any{
							"Name": map[string]any{
								"_eq": "Bob",
							},
						},
					},
				}),
				Request: `query($usersCount: Users__CountSelector) {
					_count(Users: $usersCount)
				}`,
				Results: map[string]any{
					"_count": 1,
				},
			},
		},
	}

	executeTestCase(t, test)
}

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

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"

	"github.com/sourcenetwork/immutable"
)

func TestQuerySimpleWithNonNullVariable(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21
				}`,
			},
			&action.Request{
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
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21
				}`,
			},
			&action.Request{
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
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21
				}`,
			},
			&action.Request{
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
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21
				}`,
			},
			&action.Request{
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
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21
				}`,
			},
			&action.Request{
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
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21
				}`,
			},
			&action.Request{
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
					COUNT(Users: $usersCount)
				}`,
				Results: map[string]any{
					"COUNT": 1,
				},
			},
		},
	}

	executeTestCase(t, test)
}

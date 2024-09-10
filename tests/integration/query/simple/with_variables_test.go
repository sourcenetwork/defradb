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

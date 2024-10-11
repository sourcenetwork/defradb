// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package json

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryJSON_WithGreaterThanFilterBlockWithGreaterValue_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic filter(custom), greater than",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Custom": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "David",
					"Custom": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {_gt: 20}}) {
						Name
						Custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":   "John",
							"Custom": int64(21),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithGreaterThanFilterBlockWithLesserValue_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic filter(custom), greater than",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Custom": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "David",
					"Custom": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {_gt: 22}}) {
						Name
						Custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithGreaterThanFilterBlockWithNullFilterValue_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic JSON greater than filter, with null filter value",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Custom": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "David"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {_gt: null}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithGreaterThanFilterBlockWithNestedGreaterValue_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic filter(custom), nested greater than",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Custom": {"age": 21}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "David",
					"Custom": {"age": 19}
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {age: {_gt: 20}}}) {
						Name
						Custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"Custom": map[string]any{
								"age": uint64(21),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithGreaterThanFilterBlockWithNestedLesserValue_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic filter(custom), nested greater than",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Custom": {"age": 21}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "David",
					"Custom": {"age": 19}
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {age: {_gt: 22}}}) {
						Name
						Custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithGreaterThanFilterBlockWithNestedNullFilterValue_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic JSON greater than filter, with nested null filter value",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Custom": {"age": 21}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "David"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {age: {_gt: null}}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithGreaterThanFilterBlockWithBoolValue_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic filter(custom), greater than",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Custom": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "David",
					"Custom": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {_gt: false}}) {
						Name
						Custom
					}
				}`,
				ExpectedError: `unexpected type. Property: condition, Actual: bool`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithGreaterThanFilterBlockWithStringValue_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic filter(custom), greater than",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Custom": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "David",
					"Custom": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {_gt: ""}}) {
						Name
						Custom
					}
				}`,
				ExpectedError: `unexpected type. Property: condition, Actual: string`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithGreaterThanFilterBlockWithObjectValue_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic filter(custom), greater than",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Custom": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "David",
					"Custom": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {_gt: {one: 1}}}) {
						Name
						Custom
					}
				}`,
				ExpectedError: `unexpected type. Property: condition, Actual: map[string]interface {}`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithGreaterThanFilterBlockWithArrayValue_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic filter(custom), greater than",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Custom": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "David",
					"Custom": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {_gt: [1,2]}}) {
						Name
						Custom
					}
				}`,
				ExpectedError: `unexpected type. Property: condition, Actual: []interface {}`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithGreaterThanFilterWithAllTypes_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with JSON _gt filter all types",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Shahzad",
					"Custom": "32"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Andy",
					"Custom": [1, 2]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Fred",
					"Custom": {"one": 1}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Custom": false
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "David",
					"Custom": 32
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {_gt: 30}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "David",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

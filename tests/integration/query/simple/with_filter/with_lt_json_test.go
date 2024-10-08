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
)

func TestQuerySimple_WithJSONLesserThanFilterBlockWithGreaterValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic filter(custom), lesser than",
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
					"Name": "Bob",
					"Custom": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {_lt: 20}}) {
						Name
						Custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":   "Bob",
							"Custom": int64(19),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithJSONLesserThanFilterBlockWithLesserValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic filter(custom), lesser than",
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
					"Name": "Bob",
					"Custom": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {_lt: 19}}) {
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

func TestQuerySimple_WithJSONLesserThanFilterBlockWithNullFilterValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic JSON lesser than filter, with null filter value",
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
					"Name": "Bob"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {_lt: null}}) {
						Name
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

func TestQuerySimple_WithJSONLesserThanFilterBlockWithNestedGreaterValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic filter(custom), nested lesser than",
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
					"Name": "Bob",
					"Custom": {"age": 19}
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {age: {_lt: 20}}}) {
						Name
						Custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
							"Custom": map[string]any{
								"age": uint64(19),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithJSONLesserThanFilterBlockWithNestedLesserValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic filter(custom), nested lesser than",
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
					"Name": "Bob",
					"Custom": {"age": 19}
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {age: {_lt: 19}}}) {
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

func TestQuerySimple_WithJSONLesserThanFilterBlockWithNestedNullFilterValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic JSON lesser than filter, with nested null filter value",
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
					"Name": "Bob"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {age: {_lt: null}}}) {
						Name
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

func TestQuerySimple_WithJSONLesserThanFilterBlockWithBoolValue_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic filter(custom), lesser than",
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
					"Name": "Bob",
					"Custom": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {_lt: false}}) {
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

func TestQuerySimple_WithJSONLesserThanFilterBlockWithStringValue_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic filter(custom), lesser than",
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
					"Name": "Bob",
					"Custom": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {_lt: ""}}) {
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

func TestQuerySimple_WithJSONLesserThanFilterBlockWithObjectValue_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic filter(custom), lesser than",
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
					"Name": "Bob",
					"Custom": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {_lt: {one: 1}}}) {
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

func TestQuerySimple_WithJSONLesserThanFilterBlockWithArrayValue_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic filter(custom), lesser than",
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
					"Name": "Bob",
					"Custom": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {Custom: {_lt: [1,2]}}) {
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

func TestQuerySimple_WithJSONLesserThanFilterWithAllTypes_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with JSON _lt filter all types",
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
					Users(filter: {Custom: {_lt: 33}}) {
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

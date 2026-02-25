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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

func TestQueryJSON_WithGreaterEqualFilterWithEqualValue_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Custom": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "David",
					"Custom": 32
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Custom: {_geq: 32}}) {
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

func TestQueryJSON_WithGreaterEqualFilterWithGreaterValue_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Custom": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "David",
					"Custom": 32
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Custom: {_geq: 31}}) {
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

func TestQueryJSON_WithGreaterEqualFilterWithNullValue_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Custom": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "David"
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Custom: {_geq: null}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "David",
						},
						{
							"Name": "John",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithGreaterEqualFilterWithNestedEqualValue_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Custom": {"age": 21}
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "David",
					"Custom": {"age": 32}
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Custom: {age: {_geq: 32}}}) {
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

func TestQueryJSON_WithGreaterEqualFilterWithNestedGreaterValue_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Custom": {"age": 21}
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "David",
					"Custom": {"age": 32}
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Custom: {age: {_geq: 31}}}) {
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

func TestQueryJSON_WithGreaterEqualFilterWithNestedNullValue_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Custom": {"age": 21}
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "David"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Addo",
					"Custom": {"age": null}
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Custom: {age: {_geq: null}}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "David",
						},
						{
							"Name": "John",
						},
						{
							"Name": "Addo",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithGreaterEqualFilterWithBoolValue_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Custom": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "David",
					"Custom": 32
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Custom: {_geq: true}}) {
						Name
					}
				}`,
				ExpectedError: `unexpected type. Property: condition, Actual: bool`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithGreaterEqualFilterWithStringValue_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Custom": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "David",
					"Custom": 32
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Custom: {_geq: ""}}) {
						Name
					}
				}`,
				ExpectedError: `unexpected type. Property: condition, Actual: string`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithGreaterEqualFilterWithObjectValue_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Custom": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "David",
					"Custom": 32
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Custom: {_geq: {one: 1}}}) {
						Name
					}
				}`,
				ExpectedError: `unexpected type. Property: condition, Actual: map[string]interface {}`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithGreaterEqualFilterWithArrayValue_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Custom": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "David",
					"Custom": 32
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Custom: {_geq: [1, 2]}}) {
						Name
					}
				}`,
				ExpectedError: `unexpected type. Property: condition, Actual: []interface {}`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithGreaterEqualFilterWithAllTypes_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						Name: String
						Custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Shahzad",
					"Custom": "32"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Andy",
					"Custom": [1, 2]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Fred",
					"Custom": {"one": 1}
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Custom": false
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "David",
					"Custom": 32
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Custom: {_geq: 32}}) {
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

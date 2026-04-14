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

package json

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryJSON_WithGreaterEqualFilterWithEqualValue_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "_geq filter on a JSON numeric field returns documents whose value is greater than or equal to the threshold.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
		Description: "_geq filter on a JSON numeric field excludes documents whose value is strictly less than the threshold.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
		Description: "_geq null filter on a JSON field returns all documents because every value is >= null.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
		Description: "_geq filter on a nested JSON numeric field returns documents whose nested value meets the threshold.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
		Description: "_geq filter on a nested JSON path excludes documents whose nested value is below the threshold.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
		Description: "_geq null filter on a nested JSON path returns all documents including those with null or missing nested values.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
		Description: "_geq filter with a boolean value on a JSON field returns an unexpected-type error.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
		Description: "_geq filter with a string value on a JSON field returns an unexpected-type error.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
		Description: "_geq filter with an object value on a JSON field returns an unexpected-type error.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
		Description: "_geq filter with an array value on a JSON field returns an unexpected-type error.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
		Description: "_geq filter on a JSON field matches only the numeric document when mixed JSON types are stored.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

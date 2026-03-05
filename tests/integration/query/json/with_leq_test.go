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

func TestQueryJSON_WithLesserEqualFilterWithEqualValue_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
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
					Users(filter: {Custom: {_leq: 21}}) {
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

func TestQueryJSON_WithLesserEqualFilterWithLesserValue_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
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
					Users(filter: {Custom: {_leq: 31}}) {
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

func TestQueryJSON_WithLesserEqualFilterWithNullValue_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
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
					Users(filter: {Custom: {_leq: null}}) {
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

func TestQueryJSON_WithLesserEqualFilterWithNestedEqualValue_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
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
					Users(filter: {Custom: {age: {_leq: 21}}}) {
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

func TestQueryJSON_WithLesserEqualFilterWithNestedLesserValue_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
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
					Users(filter: {Custom: {age: {_leq: 31}}}) {
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

func TestQueryJSON_WithLesserEqualFilterWithNestedNullValue_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
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
					"Name": "Addo",
					"Custom": {"age": null}
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "David"
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {Custom: {age: {_leq: null}}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Addo",
						},
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

func TestQueryJSON_WithLesserEqualFilterWithBoolValue_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
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
					Users(filter: {Custom: {_leq: true}}) {
						Name
					}
				}`,
				ExpectedError: `unexpected type. Property: condition, Actual: bool`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithLesserEqualFilterWithStringValue_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
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
					Users(filter: {Custom: {_leq: ""}}) {
						Name
					}
				}`,
				ExpectedError: `unexpected type. Property: condition, Actual: string`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithLesserEqualFilterWithObjectValue_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
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
					Users(filter: {Custom: {_leq: {one: 1}}}) {
						Name
					}
				}`,
				ExpectedError: `unexpected type. Property: condition, Actual: map[string]interface {}`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithLesserEqualFilterWithArrayValue_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
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
					Users(filter: {Custom: {_leq: [1, 2]}}) {
						Name
					}
				}`,
				ExpectedError: `unexpected type. Property: condition, Actual: []interface {}`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithLesserEqualFilterWithAllTypes_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
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
					Users(filter: {Custom: {_leq: 32}}) {
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

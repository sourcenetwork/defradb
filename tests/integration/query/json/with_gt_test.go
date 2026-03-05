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
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

func TestQueryJSON_WithGreaterThanFilterBlockWithGreaterValue_ShouldFilter(t *testing.T) {
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
					"Custom": 19
				}`,
			},
			&action.Request{
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
					"Custom": 19
				}`,
			},
			&action.Request{
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
					"Custom": {"age": 19}
				}`,
			},
			&action.Request{
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
								"age": float64(21),
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
					"Custom": {"age": 19}
				}`,
			},
			&action.Request{
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
			&action.Request{
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
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
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
					"Custom": 19
				}`,
			},
			&action.Request{
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
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
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
					"Custom": 19
				}`,
			},
			&action.Request{
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
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
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
					"Custom": 19
				}`,
			},
			&action.Request{
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
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
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
					"Custom": 19
				}`,
			},
			&action.Request{
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
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
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

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

func TestQueryJSON_WithNotEqualFilterWithObject_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"custom": {
						"tree": "maple",
						"age": 250
					}
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Andy",
					"custom": {
						"tree": "oak",
						"age": 450
					}
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": null
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {custom: {_neq: {tree:"oak",age:450}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Shahzad"},
						{"name": "John"},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithNotEqualFilterWithNestedObjects_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"custom": {
						"level_1": {
							"level_2": {
								"level_3": [true, false]
							}
						}
					}
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Andy",
					"custom": {
						"level_1": {
							"level_2": {
								"level_3": [false, true]
							}
						}
					}
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {custom: {_neq: {level_1: {level_2: {level_3: [true, false]}}}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Andy"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithNotEqualFilterWithNullValue_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"custom": null
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Andy",
					"custom": {}
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {custom: {_neq: null}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Andy"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithNeFilterAgainstNumberField_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						custom: JSON 
					}`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"age": 48,
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						"age": nil,
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Shahzad",
					"custom": map[string]any{
						"age": 42,
					},
				},
			},
			&action.Request{
				Request: `query {
					User(filter: {custom: {age: {_neq: 48}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
						{"name": "Shahzad"},
					},
				},
				NonOrderedResults: true,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithNeFilterAgainstStringField_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						custom: JSON 
					}`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"city": "Istanbul",
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						"city": nil,
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Shahzad",
					"custom": map[string]any{
						"city": "Lucerne",
					},
				},
			},
			&action.Request{
				Request: `query {
					User(filter: {custom: {city: {_neq: "Istanbul"}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
						{"name": "Andy"},
					},
				},
				NonOrderedResults: true,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithNeFilterAgainstBooleanField_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						custom: JSON 
					}`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"verified": true,
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						"verified": nil,
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Shahzad",
					"custom": map[string]any{
						"verified": false,
					},
				},
			},
			&action.Request{
				Request: `query {
					User(filter: {custom: {verified: {_neq: true}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
						{"name": "Andy"},
					},
				},
				NonOrderedResults: true,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithNeFilterAgainstNullField_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						custom: JSON 
					}`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"age": 48,
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						"age": nil,
					},
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Shahzad",
					"custom": map[string]any{
						"age": 42,
					},
				},
			},
			&action.Request{
				Request: `query {
					User(filter: {custom: {age: {_neq: null}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
						{"name": "Shahzad"},
					},
				},
				NonOrderedResults: true,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithNotEqualFilterWithNestedObject_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						custom: JSON 
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"custom": {"nested": {"foo": "bar"}}
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "David",
					"custom": {"nested": {"foo": "baz"}}
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Bruno",
					"custom": {"nested": "scalar"}
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Andy",
					"custom": {"other": {"foo": "bar"}}
				}`,
			},
			&action.Request{
				Request: `query {
					User(filter: {custom: {nested: {_neq: {foo: "bar"}}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "David"},
						{"name": "Bruno"},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

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

func TestQueryJSON_WithNotEqualFilterWithObject_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {
						"tree": "maple",
						"age": 250
					}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"custom": {
						"tree": "oak",
						"age": 450
					}
				}`,
			},
			testUtils.CreateDoc{
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
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
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
			testUtils.CreateDoc{
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
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": null
				}`,
			},
			testUtils.CreateDoc{
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
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON 
					}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"age": 48,
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						"age": nil,
					},
				},
			},
			testUtils.CreateDoc{
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
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON 
					}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"city": "Istanbul",
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						"city": nil,
					},
				},
			},
			testUtils.CreateDoc{
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
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithNeFilterAgainstBooleanField_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON 
					}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"verified": true,
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						"verified": nil,
					},
				},
			},
			testUtils.CreateDoc{
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
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithNeFilterAgainstNullField_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON 
					}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"age": 48,
					},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						"age": nil,
					},
				},
			},
			testUtils.CreateDoc{
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

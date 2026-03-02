// Copyright 2025 Democratized Data Foundation
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

func TestQueryJSON_WithEqualFilterWithObject_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
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
					Users(filter: {custom: {_eq: {tree:"oak",age:450}}}) {
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

func TestQueryJSON_WithCompoundFilterCondition_ShouldFilter(t *testing.T) {
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
					"name": "Andy",
					"custom": {
						"tree": "maple",
						"age": 450
					}
				}`,
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
					"name": "Shahzad",
					"custom": {
						"tree": "maple",
						"age": 20
					}
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {_and: [
						{custom: {tree: {_eq: "maple"}}},
						{custom: {age: {_eq: 250}}}
					]}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "John"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithEqualFilterWithNestedObjects_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
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
					Users(filter: {custom: {_eq: {level_1: {level_2: {level_3: [true, false]}}}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "John"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithEqualFilterWithNullValue_ShouldFilter(t *testing.T) {
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
					Users(filter: {custom: {_eq: null}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "John"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithEqualFilterWithAllTypes_ShouldFilter(t *testing.T) {
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
					Users(filter: {Custom: {_eq: {one: 1}}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Fred",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithEqualFilterWithObjectValueOnNestedPath_ShouldFilter(t *testing.T) {
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
					User(filter: {custom: {nested: {_eq: {foo: "bar"}}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

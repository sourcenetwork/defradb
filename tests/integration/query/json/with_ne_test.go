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

func TestQueryJSON_WithNotEqualFilterWithObject_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
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
			testUtils.Request{
				Request: `query {
					Users(filter: {custom: {_ne: {tree:"oak",age:450}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "John"},
						{"name": "Shahzad"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryJSON_WithNotEqualFilterWithNestedObjects_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
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
			testUtils.Request{
				Request: `query {
					Users(filter: {custom: {_ne: {level_1: {level_2: {level_3: [true, false]}}}}}) {
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
			testUtils.SchemaUpdate{
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
			testUtils.Request{
				Request: `query {
					Users(filter: {custom: {_ne: null}}) {
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

func TestQueryJSON_WithNeFilterAgainstNonNullValue_ShouldFetchNullValues(t *testing.T) {
	type testCase struct {
		name   string
		req    string
		result map[string]any
	}

	testCases := []testCase{
		{
			name: "query number field",
			req: `query {
				User(filter: {custom: {age: {_ne: 48}}}) {
					name
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "Shahzad"},
					{"name": "Andy"},
				},
			},
		},
		{
			name: "query string field",
			req: `query {
				User(filter: {custom: {city: {_ne: "Istanbul"}}}) {
					name	
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "Shahzad"},
					{"name": "Andy"},
				},
			},
		},
		{
			name: "query bool field",
			req: `query {
				User(filter: {custom: {verified: {_ne: true}}}) {
					name	
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "Shahzad"},
					{"name": "Andy"},
				},
			},
		},
		{
			name: "query null field",
			req: `query {
				User(filter: {custom: {age: {_ne: null}}}) {
					name	
				}
			}`,
			result: map[string]any{
				"User": []map[string]any{
					{"name": "Shahzad"},
					{"name": "John"},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			test := testUtils.TestCase{
				Actions: []any{
					testUtils.SchemaUpdate{
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
								"age":      48,
								"city":     "Istanbul",
								"verified": true,
							},
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Andy",
							"custom": map[string]any{
								"age":      nil,
								"city":     nil,
								"verified": nil,
							},
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Shahzad",
							"custom": map[string]any{
								"age":      42,
								"city":     "Lucerne",
								"verified": false,
							},
						},
					},
					testUtils.CreateDoc{
						DocMap: map[string]any{
							"name": "Fred",
							"custom": map[string]any{
								"other": "value",
							},
						},
					},
					testUtils.Request{
						Request: tc.req,
						Results: tc.result,
					},
				},
			}

			testUtils.ExecuteTestCase(t, test)
		})
	}
}

// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package index

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestJSONIndex_WithFilterOnNumberField_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {height: {_eq: 168}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {"height": 168, "weight": 70}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"height": 168}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": {"weight": 80, "BMI": 25}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": {"height": "168 cm"}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Bruno",
					"custom": {"height": null}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"custom": {"height": 190}
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
						{"name": "Islam"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(2).WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithFilterOnStringField_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {title: {_eq: "Mr"}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {"title": "Mr", "weight": 70}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"title": "Mr"}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": {"weight": 80, "BMI": 25}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": {"title": 7}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Bruno",
					"custom": {"title": null}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"custom": {"title": "Dr"}
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
						{"name": "Islam"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(2).WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithFilterOnBoolField_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {isStudent: {_eq: true}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {"isStudent": true, "weight": 70}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"isStudent": true}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": {"weight": 80, "BMI": 25}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": {"isStudent": "very much true"}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Bruno",
					"custom": {"isStudent": null}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"custom": {"isStudent": false}
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Islam"},
						{"name": "John"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(2).WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithEqFilterOnNullField_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {title: {_eq: null}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {"title": null, "weight": 70}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"title": null}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": {"weight": 80, "BMI": 25}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": {"title": "null"}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Bruno",
					"custom": {"title": 0}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"custom": {"title": "Dr"}
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
						{"name": "Islam"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(2).WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}


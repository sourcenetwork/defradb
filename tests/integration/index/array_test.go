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

func TestArrayIndex_WithFilterOnIndexedArrayUsingAny_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {numbers: {_any: {_eq: 30}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 10, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [30, 40, 50, 30]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [33, 44, 55]
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(2).WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayIndex_WithFilterOnIndexedArrayUsingAll_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {numbers: {_all: {_ge: 33}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 10, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [30, 40, 50]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [33, 44, 55]
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Andy",
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(9),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayIndex_WithFilterOnIndexedArrayUsingNone_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {numbers: {_none: {_ge: 33}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 10, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [30, 40, 50]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [33, 44, 55]
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(9),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayIndexUpdate_IfUpdateRearrangesArrayElements_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {numbers: {_any: {_eq: 30}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 10, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [30, 40, 50, 30]
				}`,
			},
			testUtils.UpdateDoc{
				DocID: 1,
				Doc: `{
					"name": "Shahzad",
					"numbers": [50, 30, 40]
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(2).WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayIndexUpdate_IfUpdateRemovesSoughtElement_ShouldNotFetch(t *testing.T) {
	req := `query {
		User(filter: {numbers: {_any: {_eq: 30}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 10, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [30, 40, 50, 30]
				}`,
			},
			testUtils.UpdateDoc{
				DocID: 1,
				Doc: `{
					"name": "Shahzad",
					"numbers": [50, 40]
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayIndexUpdate_IfUpdateAddsSoughtElement_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {numbers: {_any: {_eq: 30}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 10, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [40, 50]
				}`,
			},
			testUtils.UpdateDoc{
				DocID: 1,
				Doc: `{
					"name": "Shahzad",
					"numbers": [80, 30, 60]
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayIndexDelete_IfUpdateRemovesSoughtElement_ShouldNotFetch(t *testing.T) {
	req := `query {
		User(filter: {numbers: {_any: {_gt: 0}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 10, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [40, 50]
				}`,
			},
			testUtils.DeleteDoc{DocID: 0},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
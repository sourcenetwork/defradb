// Copyright 2025 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryWithIndex_WithEqFilterOnDateTimeField_ShouldIndex(t *testing.T) {
	req := `query {
		User(filter: {birthday: {_eq: "2000-07-23T03:00:00-00:00"}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						birthday: DateTime @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"name":	"Fred",
						"birthday": "2000-07-23T03:00:00-00:00"
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"name":	"Andy",
						"birthday": "2001-08-23T03:00:00-00:00"
					}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Fred"},
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

func TestQueryWithIndex_WithGtFilterOnDateTimeField_ShouldIndex(t *testing.T) {
	req := `query {
		User(filter: {birthday: {_gt: "2001-08-23T03:00:00-00:00"}}) {
			name
		}
	}`

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						birthday: DateTime @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"name":	"Shahzad",
						"birthday": "2001-08-24T03:00:00-00:00"
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"name":	"Fred",
						"birthday": "2000-08-22T03:00:00-00:00"
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"name":	"Andy",
						"birthday": "2001-08-23T03:00:00-00:00"
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"name":	"John",
						"birthday": "2001-08-25T03:00:00-00:00"
					}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
						{"name": "John"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithGeFilterOnDateTimeField_ShouldIndex(t *testing.T) {
	req := `query {
		User(filter: {birthday: {_ge: "2001-01-01T00:00:00-00:00"}}) {
			name
		}
	}`

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						birthday: DateTime @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"name":	"Fred",
						"birthday": "2000-07-23T03:00:00-00:00"
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"name":	"Andy",
						"birthday": "2001-08-23T03:00:00-00:00"
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"name":	"Keenan",
						"birthday": "2001-01-01T00:00:00-00:00"
					}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Keenan"},
						{"name": "Andy"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithLtFilterOnDateTimeField_ShouldIndex(t *testing.T) {
	req := `query {
		User(filter: {birthday: {_lt: "2001-01-01T00:00:00-00:00"}}) {
			name
		}
	}`

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						birthday: DateTime @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"name":	"Fred",
						"birthday": "2000-07-23T03:00:00-00:00"
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"name":	"Andy",
						"birthday": "2001-08-23T03:00:00-00:00"
					}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Fred"},
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

func TestQueryWithIndex_WithLeFilterOnDateTimeField_ShouldIndex(t *testing.T) {
	req := `query {
		User(filter: {birthday: {_le: "2001-01-01T00:00:00-00:00"}}) {
			name
		}
	}`

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						birthday: DateTime @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"name":	"Fred",
						"birthday": "2000-07-23T03:00:00-00:00"
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"name":	"Andy",
						"birthday": "2001-08-23T03:00:00-00:00"
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"name":	"Keenan",
						"birthday": "2001-01-01T00:00:00-00:00"
					}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Fred"},
						{"name": "Keenan"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithNeFilterOnDateTimeField_ShouldIndex(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						birthday: DateTime @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"name":	"Fred",
						"birthday": "2000-07-23T03:00:00-00:00"
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"name":	"Andy",
						"birthday": "2001-08-23T03:00:00-00:00"
					}`,
			},
			testUtils.Request{
				Request: `query {
					User(filter: {birthday: {_ne: "2000-07-23T03:00:00-00:00"}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

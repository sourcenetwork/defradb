// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimple_WithGroupAverageAliasFilter_FiltersResults(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group average alias filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `type Users {
					Name: String
					Score: Int
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Score": 10
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Score": 20
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Score": 40
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Score": 0
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Name], filter: {_alias: {averageScore: {_eq: 20}}}) {
						Name
						averageScore: _avg(_group: {field: Score})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":         "Alice",
							"averageScore": float64(20),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithGroupSumAliasFilter_FiltersResults(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group sum alias filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `type Users {
					Name: String
					Score: Int
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Score": 10
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Score": 20
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Score": 40
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Score": 0
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Name], filter: {_alias: {totalScore: {_eq: 40}}}) {
						Name
						totalScore: _sum(_group: {field: Score})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":       "Alice",
							"totalScore": float64(40),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithGroupMinAliasFilter_FiltersResults(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group min alias filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `type Users {
					Name: String
					Score: Int
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Score": 10
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Score": 20
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Score": 40
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Score": 0
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Name], filter: {_alias: {minScore: {_eq: 0}}}) {
						Name
						minScore: _min(_group: {field: Score})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":     "Alice",
							"minScore": int64(0),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithGroupMaxAliasFilter_FiltersResults(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group max alias filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `type Users {
					Name: String
					Score: Int
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Score": 10
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Score": 20
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Score": 40
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Score": 0
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Name], filter: {_alias: {maxScore: {_eq: 40}}}) {
						Name
						maxScore: _max(_group: {field: Score})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":     "Alice",
							"maxScore": int64(40),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithGroupCountAliasFilter_FiltersResults(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group count alias filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `type Users {
					Name: String
					Score: Int
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Score": 10
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Score": 20
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Score": 40
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Score": 0
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Score": 5
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Name], filter: {_alias: {scores: {_eq: 3}}}) {
						Name
						scores: _count(_group: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":   "Alice",
							"scores": int64(3),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

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

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimple_WithGroupAverageAliasFilter_FiltersResults(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users {
					Name: String
					Score: Int
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Score": 10
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Score": 20
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Score": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Score": 0
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name], filter: {_alias: {averageScore: {_eq: 20}}}) {
						Name
						averageScore: AVG(GROUP: {field: Score})
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
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users {
					Name: String
					Score: Int
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Score": 10
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Score": 20
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Score": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Score": 0
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name], filter: {_alias: {totalScore: {_eq: 40}}}) {
						Name
						totalScore: SUM(GROUP: {field: Score})
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
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users {
					Name: String
					Score: Int
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Score": 10
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Score": 20
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Score": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Score": 0
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name], filter: {_alias: {minScore: {_eq: 0}}}) {
						Name
						minScore: MIN(GROUP: {field: Score})
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
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users {
					Name: String
					Score: Int
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Score": 10
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Score": 20
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Score": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Score": 0
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name], filter: {_alias: {maxScore: {_eq: 40}}}) {
						Name
						maxScore: MAX(GROUP: {field: Score})
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
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users {
					Name: String
					Score: Int
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Score": 10
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Score": 20
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Score": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Score": 0
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Score": 5
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name], filter: {_alias: {scores: {_eq: 3}}}) {
						Name
						scores: COUNT(GROUP: {})
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

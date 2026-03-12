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

package index

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestArrayUniqueCompositeIndex_WithUniqueCombinations_Succeed(t *testing.T) {
	req := `query {
		User(filter: {nfts1: {_any: {_eq: 2}}, nfts2: {_any: {_eq: 3}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User @index(unique: true, includes: [{field: "nfts1"}, {field: "nfts2"}]) {
						name: String 
						nfts1: [Int!] 
						nfts2: [Int!] 
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"nfts1": [1, 2],
					"nfts2": [1, 3]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"nfts1": [1, 2],
					"nfts2": [2, 4]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Keenan",
					"nfts1": [3, 4],
					"nfts2": [1, 3]
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayUniqueCompositeIndex_IfDocIsAddedThatViolatesUniqueness_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User @index(unique: true, includes: [{field: "nfts1"}, {field: "nfts2"}]) {
						name: String 
						nfts1: [Int!] 
						nfts2: [Int!] 
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"nfts1": [1, 2],
					"nfts2": [1, 3]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"nfts1": [1, 2],
					"nfts2": [2, 4, 3]
				}`,
				ExpectedError: "can not index a doc's field(s) that violates unique index.",
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"nfts1": [5, 6, 2],
					"nfts2": [1, 3]
				}`,
				ExpectedError: "can not index a doc's field(s) that violates unique index.",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayUniqueCompositeIndex_IfDocIsUpdatedThatViolatesUniqueness_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User @index(unique: true, includes: [{field: "nfts1"}, {field: "nfts2"}]) {
						name: String 
						nfts1: [Int!] 
						nfts2: [Int!] 
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"nfts1": [1, 2],
					"nfts2": [1, 3]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"nfts1": [1, 2],
					"nfts2": [2, 4, 5, 6]
				}`,
			},
			testUtils.UpdateDoc{
				DocID: 1,
				Doc: `{
					"name": "Shahzad",
					"nfts1": [1],
					"nfts2": [2, 5, 3]
				}`,
				ExpectedError: "can not index a doc's field(s) that violates unique index.",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayUniqueCompositeIndex_IfDocsHaveNilValues_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User @index(unique: true, includes: [{field: "nfts1"}, {field: "nfts2"}]) {
						name: String 
						nfts1: [Int] 
						nfts2: [Int] 
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"nfts1": [1, null],
					"nfts2": [null, 1, 3, null]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"nfts1": [1, null, 2],
					"nfts2": [2, 4, null, 5, 6, null]
				}`,
			},
			&action.Request{
				Request: `query {
						User(filter: {nfts1: {_any: {_eq: null}}, nfts2: {_any: {_eq: null}}}) {
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

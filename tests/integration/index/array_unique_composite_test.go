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

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/db"
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
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "nfts1"}, {field: "nfts2"}]) {
						name: String 
						nfts1: [Int!] 
						nfts2: [Int!] 
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"nfts1": [1, 2],
					"nfts2": [1, 3]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"nfts1": [1, 2],
					"nfts2": [2, 4]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"nfts1": [3, 4],
					"nfts2": [1, 3]
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
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

func TestArrayUniqueCompositeIndex_IfDocIsCreatedThatViolatesUniqueness_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "nfts1"}, {field: "nfts2"}]) {
						name: String 
						nfts1: [Int!] 
						nfts2: [Int!] 
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"nfts1": [1, 2],
					"nfts2": [1, 3]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"nfts1": [1, 2],
					"nfts2": [2, 4, 3]
				}`,
				ExpectedError: db.NewErrCanNotIndexNonUniqueFields(
					"bae-02823b81-729a-5cb8-88cb-6df2e15232b1",
					errors.NewKV("nfts1", []int64{1, 2}), errors.NewKV("nfts2", []int64{2, 4, 3})).Error(),
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"nfts1": [5, 6, 2],
					"nfts2": [1, 3]
				}`,
				ExpectedError: db.NewErrCanNotIndexNonUniqueFields(
					"bae-ccb3fd51-caf9-5b34-b2d2-e4ad020409e1",
					errors.NewKV("nfts1", []int64{5, 6, 2}), errors.NewKV("nfts2", []int64{1, 3})).Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayUniqueCompositeIndex_IfDocIsUpdatedThatViolatesUniqueness_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "nfts1"}, {field: "nfts2"}]) {
						name: String 
						nfts1: [Int!] 
						nfts2: [Int!] 
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"nfts1": [1, 2],
					"nfts2": [1, 3]
				}`,
			},
			testUtils.CreateDoc{
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
				ExpectedError: db.NewErrCanNotIndexNonUniqueFields(
					"bae-f6b3ab5a-dfa4-53fd-a320-a3e203a9e6f5",
					errors.NewKV("nfts1", []int64{1}), errors.NewKV("nfts2", []int64{2, 5, 3})).Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayUniqueCompositeIndex_IfDocsHaveNilValues_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "nfts1"}, {field: "nfts2"}]) {
						name: String 
						nfts1: [Int] 
						nfts2: [Int] 
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"nfts1": [1, null],
					"nfts2": [null, 1, 3, null]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"nfts1": [1, null, 2],
					"nfts2": [2, 4, null, 5, 6, null]
				}`,
			},
			testUtils.Request{
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

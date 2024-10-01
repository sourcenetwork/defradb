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

func TestArrayCompositeUniqueIndex_UponDocCreationWithArrayElementThatExists_Error(t *testing.T) {
	req := `query {
		User(filter: {nfts: {_any: {_eq: 30}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						nfts: [Int!] @index(unique: true)
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"nfts": [0, 30, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"nfts": [10, 40]
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

func TestArrayCompositeUniqueIndex_UponDocCreationWithUniqueElements_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						nfts: [Int!] @index(unique: true)
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"nfts": [0, 30, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"nfts": [50, 30]
				}`,
				ExpectedError: db.NewErrCanNotIndexNonUniqueFields(
					"bae-a4045a20-b9e6-5b19-82d5-5e54176895a8",
					errors.NewKV("nfts", []int64{50, 30})).Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayCompositeUniqueIndex_UponDocUpdateWithUniqueElements_Succeed(t *testing.T) {
	req := `query {
		User(filter: {nfts: {_any: {_eq: 60}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						nfts: [Int!] @index(unique: true)
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"nfts": [0, 30, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"nfts": [10, 40]
				}`,
			},
			testUtils.UpdateDoc{
				DocID: 1,
				Doc: `{
					"nfts": [10, 60]
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
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

func TestArrayCompositeUniqueIndex_UponDocUpdateWithArrayElementThatExists_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User  {
						name: String 
						nfts: [Int!] @index(unique: true)
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"nfts": [0, 30, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"nfts": [10, 40]
				}`,
			},
			testUtils.UpdateDoc{
				DocID: 1,
				Doc: `{
					"nfts": [50, 30]
				}`,
				ExpectedError: db.NewErrCanNotIndexNonUniqueFields(
					"bae-d065234c-4bf5-5cb8-8068-6f1fda8ed661",
					errors.NewKV("nfts", []int64{50, 30})).Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayCompositeUniqueIndex_UponDeletingDoc_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User  {
						name: String 
						nfts: [Int!] @index(unique: true)
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"nfts": [0, 30, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"nfts": [10, 40]
				}`,
			},
			testUtils.DeleteDoc{
				DocID: 1,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

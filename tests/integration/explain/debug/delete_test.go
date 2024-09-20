// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain_debug

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

var deletePattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"deleteNode": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"scanNode": dataMap{},
						},
					},
				},
			},
		},
	},
}

func TestDebugExplainMutationRequestWithDeleteUsingFilter(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) mutation request with delete using filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain(type: debug) {
					delete_Author(filter: {name: {_eq: "Shahzad"}}) {
						_docID
					}
				}`,

				ExpectedPatterns: deletePattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainMutationRequestWithDeleteUsingFilterToMatchEverything(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) mutation request with delete using filter to match everything.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain(type: debug) {
					delete_Author(filter: {}) {
						DeletedKeyByFilter: _docID
					}
				}`,

				ExpectedPatterns: deletePattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainMutationRequestWithDeleteUsingId(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) mutation request with delete using document id.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain(type: debug) {
					delete_Author(docID: "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d") {
						_docID
					}
				}`,

				ExpectedPatterns: deletePattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainMutationRequestWithDeleteUsingIds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) mutation request with delete using ids.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain(type: debug) {
					delete_Author(docID: [
						"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
						"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f"
					]) {
						AliasKey: _docID
					}
				}`,

				ExpectedPatterns: deletePattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainMutationRequestWithDeleteUsingNoIds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) mutation request with delete using no ids.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain(type: debug) {
					delete_Author(docID: []) {
						_docID
					}
				}`,

				ExpectedPatterns: deletePattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainMutationRequestWithDeleteUsingFilterAndIds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) mutation request with delete using filter and ids.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain(type: debug) {
					delete_Author(
						docID: ["bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d", "test"],
						filter: {
							_and: [
								{age: {_lt: 26}},
								{verified: {_eq: true}},
							]
						}
					) {
						_docID
					}
				}`,

				ExpectedPatterns: deletePattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

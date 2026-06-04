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

package test_explain_debug

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

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

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

var updatePattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"selectTopNode": dataMap{
					"selectNode": dataMap{
						"updateNode": dataMap{
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"scanNode": dataMap{},
								},
							},
						},
					},
				},
			},
		},
	},
}

func TestDebugExplainMutationRequestWithUpdateUsingBooleanFilter(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `mutation @explain(type: debug) {
					update_Author(
						filter: {
							verified: {
								_eq: true
							}
						},
						input: {age: 59}
					) {
						_docID
						name
						age
					}
				}`,

				ExpectedPatterns: updatePattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainMutationRequestWithUpdateUsingIds(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `mutation @explain(type: debug) {
					update_Author(
						docID: [
							"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
							"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
						],
						input: {age: 59}
					) {
						_docID
						name
						age
					}
				}`,

				ExpectedPatterns: updatePattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainMutationRequestWithUpdateUsingId(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `mutation @explain(type: debug) {
					update_Author(
						docID: "bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
						input: {age: 59}
					) {
						_docID
						name
						age
					}
				}`,

				ExpectedPatterns: updatePattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainMutationRequestWithUpdateUsingIdsAndFilter(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `mutation @explain(type: debug) {
					update_Author(
						filter: {
							verified: {
								_eq: true
							}
						},
						docID: [
							"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
							"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
						],
						input: {age: 59}
					) {
						_docID
						name
						age
					}
				}`,

				ExpectedPatterns: updatePattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

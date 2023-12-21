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

var updatePattern = dataMap{
	"explain": dataMap{
		"updateNode": dataMap{
			"selectTopNode": dataMap{
				"selectNode": dataMap{
					"scanNode": dataMap{},
				},
			},
		},
	},
}

func TestDebugExplainMutationRequestWithUpdateUsingBooleanFilter(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) mutation request with update using boolean filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain(type: debug) {
					update_Author(
						filter: {
							verified: {
								_eq: true
							}
						},
						input: {age: 59}
					) {
						_key
						name
						age
					}
				}`,

				ExpectedPatterns: []dataMap{updatePattern},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainMutationRequestWithUpdateUsingIds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) mutation request with update using ids.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain(type: debug) {
					update_Author(
						ids: [
							"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
							"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
						],
						input: {age: 59}
					) {
						_key
						name
						age
					}
				}`,

				ExpectedPatterns: []dataMap{updatePattern},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainMutationRequestWithUpdateUsingId(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) mutation request with update using id.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain(type: debug) {
					update_Author(
						id: "bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
						input: {age: 59}
					) {
						_key
						name
						age
					}
				}`,

				ExpectedPatterns: []dataMap{updatePattern},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainMutationRequestWithUpdateUsingIdsAndFilter(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) mutation request with update using both ids and filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain(type: debug) {
					update_Author(
						filter: {
							verified: {
								_eq: true
							}
						},
						ids: [
							"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
							"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
						],
						input: {age: 59}
					) {
						_key
						name
						age
					}
				}`,

				ExpectedPatterns: []dataMap{updatePattern},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

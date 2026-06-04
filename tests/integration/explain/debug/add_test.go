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

var addPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"addNode": dataMap{
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

func TestDebugExplainMutationRequestWithAdd(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `mutation @explain(type: debug) {
					add_Author(input: {name: "Shahzad Lone", age: 27, verified: true}) {
						name
						age
					}
				}`,

				ExpectedPatterns: addPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainMutationRequestDoesNotAddDocGivenDuplicate(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `mutation @explain(type: debug) {
					add_Author(input: {name: "Shahzad Lone", age: 27}) {
						name
						age
					}
				}`,

				ExpectedPatterns: addPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

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

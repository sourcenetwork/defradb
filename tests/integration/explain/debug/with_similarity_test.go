// Copyright 2025 Democratized Data Foundation
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

var similarityPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"selectTopNode": dataMap{
					"similarityNode": dataMap{
						"selectNode": dataMap{
							"scanNode": dataMap{},
						},
					},
				},
			},
		},
	},
}

func TestDebugExplainRequestWith_WithSimilarity(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with similarity.",

		Actions: []any{
			&action.AddSchema{
				Schema: `type User {
					name: String
					pointsList: [Float64!]
				}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":       "John",
					"pointsList": []float64{2, 4, 1},
				},
			},

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					User {
						name
						_similarity(pointsList: {vector: [1, 2, 0]})
					}
				}`,

				ExpectedFullGraph: similarityPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

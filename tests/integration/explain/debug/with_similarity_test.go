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

		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					pointsList: [Float64!]
				}`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name":       "John",
					"pointsList": []float64{2, 4, 1},
				},
			},

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					User {
						name
						SIMILARITY(pointsList: {vector: [1, 2, 0]})
					}
				}`,

				ExpectedFullGraph: similarityPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

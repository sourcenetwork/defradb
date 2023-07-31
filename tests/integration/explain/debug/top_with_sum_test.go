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

var topLevelSumPattern = dataMap{
	"explain": dataMap{
		"topLevelNode": []dataMap{
			{
				"selectTopNode": dataMap{
					"selectNode": dataMap{
						"scanNode": dataMap{},
					},
				},
			},
			{
				"sumNode": dataMap{},
			},
		},
	},
}

func TestDebugExplainTopLevelSumRequest(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) top-level sum request.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					_sum(
						Author: {
							field: age
						}
					)
				}`,

				ExpectedPatterns: []dataMap{topLevelSumPattern},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainTopLevelSumRequestWithFilter(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) top-level sum request with filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					_sum(
						Author: {
							field: age,
							filter: {
								age: {
									_gt: 26
								}
							}
						}
					)
				}`,

				ExpectedPatterns: []dataMap{topLevelSumPattern},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

// Copyright 2024 Democratized Data Foundation
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

var topLevelMaxPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"topLevelNode": []dataMap{
					{
						"selectTopNode": dataMap{
							"selectNode": dataMap{
								"scanNode": dataMap{},
							},
						},
					},
					{
						"maxNode": dataMap{},
					},
				},
			},
		},
	},
}

func TestDebugExplain_TopLevelMaxRequest_Succeeds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) top-level max request.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					_max(
						Author: {
							field: age
						}
					)
				}`,

				ExpectedPatterns: topLevelMaxPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplain_TopLevelMaxRequestWithFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) top-level max request with filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					_max(
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

				ExpectedPatterns: topLevelMaxPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

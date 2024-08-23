// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain_default

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

var limitPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"selectTopNode": dataMap{
					"limitNode": dataMap{
						"selectNode": dataMap{
							"scanNode": dataMap{},
						},
					},
				},
			},
		},
	},
}

func TestDefaultExplainRequestWithOnlyLimit(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with only limit.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(limit: 2) {
						name
					}
				}`,

				ExpectedPatterns: limitPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "limitNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"limit":  uint64(2),
							"offset": uint64(0),
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainRequestWithOnlyOffset(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with only offset.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(offset: 2) {
						name
					}
				}`,

				ExpectedPatterns: limitPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "limitNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"limit":  nil,
							"offset": uint64(2),
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainRequestWithLimitAndOffset(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with limit and offset.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(limit: 3, offset: 1) {
						name
					}
				}`,

				ExpectedPatterns: limitPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "limitNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"limit":  uint64(3),
							"offset": uint64(1),
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

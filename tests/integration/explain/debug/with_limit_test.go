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

func TestDebugExplainRequestWithOnlyLimit(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with only limit.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(limit: 2) {
						name
					}
				}`,

				ExpectedPatterns: limitPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithOnlyOffset(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with only offset.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(offset: 2) {
						name
					}
				}`,

				ExpectedPatterns: limitPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithLimitAndOffset(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with limit and offset.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(limit: 3, offset: 1) {
						name
					}
				}`,

				ExpectedPatterns: limitPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

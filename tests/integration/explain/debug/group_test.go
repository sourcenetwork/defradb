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

var groupPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"selectTopNode": dataMap{
					"groupNode": dataMap{
						"selectNode": dataMap{
							"pipeNode": dataMap{
								"scanNode": dataMap{},
							},
						},
					},
				},
			},
		},
	},
}

func TestDebugExplainRequestWithGroupByOnParent(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with group-by on parent.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author (groupBy: [age]) {
						age
						_group {
							name
						}
					}
				}`,

				ExpectedFullGraph: groupPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithGroupByTwoFieldsOnParent(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with group-by two fields on parent.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author (groupBy: [age, name]) {
						age
						_group {
							name
						}
					}
				}`,

				ExpectedFullGraph: groupPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

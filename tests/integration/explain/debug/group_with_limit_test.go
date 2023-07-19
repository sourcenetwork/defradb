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

var debugGroupLimitPattern = dataMap{
	"explain": dataMap{
		"selectTopNode": dataMap{
			"limitNode": dataMap{
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
}

func TestDebugExplainRequestWithLimitAndOffsetOnParentGroupBy(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with limit and offset on parent groupBy.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(
						groupBy: [name],
						limit: 1,
						offset: 1
					) {
						name
						_group {
							age
						}
					}
				}`,

				ExpectedPatterns: []dataMap{debugGroupLimitPattern},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithLimitOnParentGroupByAndInnerGroupSelection(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with limit and offset on parent groupBy and inner _group selection.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(
						groupBy: [name],
						limit: 1
					) {
						name
						_group(limit: 2) {
							age
						}
					}
				}`,

				ExpectedPatterns: []dataMap{debugGroupLimitPattern},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

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

var debugGroupOrderPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"selectTopNode": dataMap{
					"orderNode": dataMap{
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
	},
}

func TestDebugExplainRequestWithDescendingOrderOnParentGroupBy(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with order (descending) on parent groupBy.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(
						groupBy: [name],
						order: {name: DESC}
					) {
						name
						_group {
							age
						}
					}
				}`,

				ExpectedPatterns: debugGroupOrderPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithAscendingOrderOnParentGroupBy(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with order (ascending) on parent groupBy.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(
						groupBy: [name],
						order: {name: ASC}
					) {
						name
						_group {
							age
						}
					}
				}`,

				ExpectedPatterns: debugGroupOrderPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithOrderOnParentGroupByAndOnInnerGroupSelection(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with order on parent groupBy and inner _group selection.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(
						groupBy: [name],
						order: {name: DESC}
					) {
						name
						_group (order: {age: ASC}){
							age
						}
					}
				}`,

				ExpectedPatterns: debugGroupOrderPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(
						groupBy: [name],
						order: {name: DESC}
					) {
						name
						GROUP {
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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(
						groupBy: [name],
						order: {name: ASC}
					) {
						name
						GROUP {
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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(
						groupBy: [name],
						order: {name: DESC}
					) {
						name
						GROUP (order: {age: ASC}){
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

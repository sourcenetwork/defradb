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

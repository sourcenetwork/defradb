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

var debugGroupAveragePattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"selectTopNode": dataMap{
					"averageNode": dataMap{
						"countNode": dataMap{
							"sumNode": dataMap{
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
		},
	},
}

func TestDebugExplainRequestWithGroupByWithAverageOnAnInnerField(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author (groupBy: [name]) {
						name
						AVG(GROUP: {field: age})
					}
				}`,

				ExpectedPatterns: debugGroupAveragePattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithAverageInsideTheInnerGroupOnAField(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author (groupBy: [name]) {
						name
						AVG(GROUP: {field: AVG})
						GROUP(groupBy: [verified]) {
							verified
							AVG(GROUP: {field: age})
						}
					}
				}`,

				ExpectedPatterns: debugGroupAveragePattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithAverageInsideTheInnerGroupOnAFieldAndNestedGroupBy(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author (groupBy: [name]) {
						name
						AVG(GROUP: {field: AVG})
						GROUP(groupBy: [verified]) {
							verified
								AVG(GROUP: {field: age})
								GROUP (groupBy: [age]){
									age
								}
						}
					}
				}`,

				ExpectedPatterns: debugGroupAveragePattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithAverageInsideTheInnerGroupAndNestedGroupByWithAverage(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author (groupBy: [name]) {
						name
						AVG(GROUP: {field: AVG})
						GROUP(groupBy: [verified]) {
							verified
								AVG(GROUP: {field: age})
								GROUP (groupBy: [age]){
									age
									AVG(GROUP: {field: age})
								}
						}
					}
				}`,

				ExpectedPatterns: debugGroupAveragePattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

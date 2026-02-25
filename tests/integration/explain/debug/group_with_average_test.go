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

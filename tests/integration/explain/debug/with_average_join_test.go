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

var debugAverageTypeIndexJoinManyPattern = dataMap{
	"explain": dataMap{
		"selectTopNode": dataMap{
			"averageNode": dataMap{
				"countNode": dataMap{
					"sumNode": dataMap{
						"selectNode": dataMap{
							"typeIndexJoin": dataMap{
								"typeJoinMany": normalTypeJoinPattern,
							},
						},
					},
				},
			},
		},
	},
}

func TestDebugExplainRequestWithAverageOnJoinedField(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with average on joined/related field.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author {
						name
						_avg(books: {field: pages})
					}
				}`,

				ExpectedPatterns: []dataMap{debugAverageTypeIndexJoinManyPattern},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithAverageOnMultipleJoinedFieldsWithFilter(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with average on multiple joined fields with filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author {
						name
						_avg(
							books: {field: pages},
							articles: {field: pages, filter: {pages: {_gt: 3}}}
						)
					}
				}`,

				ExpectedPatterns: []dataMap{
					{
						"explain": dataMap{
							"selectTopNode": dataMap{
								"averageNode": dataMap{
									"countNode": dataMap{
										"sumNode": dataMap{
											"selectNode": dataMap{
												"parallelNode": []dataMap{
													{
														"typeIndexJoin": dataMap{
															"typeJoinMany": debugTypeJoinPattern,
														},
													},
													{
														"typeIndexJoin": dataMap{
															"typeJoinMany": debugTypeJoinPattern,
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
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

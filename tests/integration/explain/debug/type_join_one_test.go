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

func TestDebugExplainRequestWithAOneToOneJoin(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with a 1-to-1 join.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author {
						OnlyEmail: contact {
							email
						}
					}
				}`,

				ExpectedPatterns: []dataMap{
					{
						"explain": dataMap{
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"typeIndexJoin": dataMap{
										"typeJoinOne": normalTypeJoinPattern,
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

func TestDebugExplainRequestWithTwoLevelDeepNestedJoins(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with two level deep nested joins.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author {
						name
						contact {
							email
							address {
								city
							}
						}
					}
				}`,

				ExpectedPatterns: []dataMap{
					{
						"explain": dataMap{
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"typeIndexJoin": dataMap{
										"typeJoinOne": dataMap{
											"root": dataMap{
												"scanNode": dataMap{},
											},
											"subType": dataMap{
												"selectTopNode": dataMap{
													"selectNode": dataMap{
														"typeIndexJoin": dataMap{
															"typeJoinOne": normalTypeJoinPattern,
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

// Copyright 2024 Democratized Data Foundation
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

var debugMaxTypeIndexJoinManyPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"selectTopNode": dataMap{
					"maxNode": dataMap{
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

func TestDebugExplainRequest_WithMaxOnOneToManyJoinedField_Succeeds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with max on a one-to-many joined field.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author {
						name
						_docID
						MaxPages: _max(
							books: {field: pages}
						)
					}
				}`,

				ExpectedPatterns: debugMaxTypeIndexJoinManyPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequest_WithMaxOnOneToManyJoinedFieldWithFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with max on a one-to-many joined field, with filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author {
						name
						MaxPages: _max(
							articles: {
								field: pages,
								filter: {
									name: {
										_eq: "To my dear readers"
									}
								}
							}
						)
					}
				}`,

				ExpectedPatterns: debugMaxTypeIndexJoinManyPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequest_WithMaxOnOneToManyJoinedFieldWithManySources_Succeeds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with max on a one-to-many joined field with many sources.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author {
						name
						MaxPages: _max(
							books: {field: pages},
							articles: {field: pages}
						)
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"maxNode": dataMap{
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
	}

	explainUtils.ExecuteTestCase(t, test)
}

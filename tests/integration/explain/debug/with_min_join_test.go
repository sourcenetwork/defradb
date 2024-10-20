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

var debugMinTypeIndexJoinManyPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"selectTopNode": dataMap{
					"minNode": dataMap{
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

func TestDebugExplainRequest_WithMinOnOneToManyJoinedField_Succeeds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with min on a one-to-many joined field.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author {
						name
						_docID
						MinPages: _min(
							books: {field: pages}
						)
					}
				}`,

				ExpectedPatterns: debugMinTypeIndexJoinManyPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequest_WithMinOnOneToManyJoinedFieldWithFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with min on a one-to-many joined field, with filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author {
						name
						MinPages: _min(
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

				ExpectedPatterns: debugMinTypeIndexJoinManyPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequest_WithMinOnOneToManyJoinedFieldWithManySources_Succeeds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with min on a one-to-many joined field with many sources.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author {
						name
						MinPages: _min(
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
									"minNode": dataMap{
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

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

var debugAverageTypeIndexJoinManyPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
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
		},
	},
}

func TestDebugExplainRequestWithAverageOnJoinedField(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author {
						name
						AVG(books: {field: pages})
					}
				}`,

				ExpectedPatterns: debugAverageTypeIndexJoinManyPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithAverageOnMultipleJoinedFieldsWithFilter(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author {
						name
						AVG(
							books: {field: pages},
							articles: {field: pages, filter: {pages: {_gt: 3}}}
						)
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
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
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

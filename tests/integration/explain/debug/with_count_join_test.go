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

var debugCountTypeIndexJoinManyPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"selectTopNode": dataMap{
					"countNode": dataMap{
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

func TestDebugExplainRequestWithCountOnOneToManyJoinedField(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author {
						name
						numberOfBooks: COUNT(books: {})
					}
				}`,

				ExpectedPatterns: debugCountTypeIndexJoinManyPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithCountOnOneToManyJoinedFieldWithManySources(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author {
						name
						numberOfBooks: COUNT(
							books: {}
							articles: {}
						)
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"countNode": dataMap{
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

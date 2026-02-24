// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain_execute

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

func TestExecuteExplainRequestWithSumOfInlineArrayField(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			add2AddressDocuments(),
			add2AuthorContactDocuments(),
			add2AuthorDocuments(),
			add3BookDocuments(),

			&action.ExplainRequest{
				Request: `query @explain(type: execute) {
					Book {
						name
						NotSureWhySomeoneWouldSumTheChapterPagesButHereItIs: SUM(chapterPages: {})
					}
				}`,

				ExpectedFullGraph: dataMap{
					"explain": dataMap{
						"executionSuccess": true,
						"sizeOfResult":     1,
						"planExecutions":   uint64(2),
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"sumNode": dataMap{
										"iterations": uint64(4),
										"selectNode": dataMap{
											"iterations":    uint64(4),
											"filterMatches": uint64(3),
											"scanNode": dataMap{
												"iterations":   uint64(4),
												"docFetches":   uint64(3),
												"fieldFetches": uint64(11),
												"indexFetches": uint64(0),
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

func TestExecuteExplainRequestSumOfRelatedOneToManyField(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,
			add2AddressDocuments(),
			add2AuthorContactDocuments(),
			add2AuthorDocuments(),
			add3ArticleDocuments(),

			&action.ExplainRequest{
				Request: `query @explain(type: execute) {
					Author {
						name
						TotalPages: SUM(
							articles: {
								field: pages,
							}
						)
					}
				}`,

				ExpectedFullGraph: dataMap{
					"explain": dataMap{
						"executionSuccess": true,
						"sizeOfResult":     1,
						"planExecutions":   uint64(2),
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"sumNode": dataMap{
										"iterations": uint64(3),
										"selectNode": dataMap{
											"iterations":    uint64(3),
											"filterMatches": uint64(2),
											"typeIndexJoin": dataMap{
												"iterations": uint64(3),
												"typeJoinMany": dataMap{
													"root": dataMap{
														"scanNode": dataMap{
															"iterations":   uint64(3),
															"docFetches":   uint64(2),
															"fieldFetches": uint64(8),
															"indexFetches": uint64(0),
														},
													},
													"subType": dataMap{
														"selectTopNode": dataMap{
															"selectNode": dataMap{
																"iterations":    uint64(5),
																"filterMatches": uint64(3),
																"scanNode": dataMap{
																	"iterations":   uint64(5),
																	"docFetches":   uint64(6),
																	"fieldFetches": uint64(18),
																	"indexFetches": uint64(0),
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
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

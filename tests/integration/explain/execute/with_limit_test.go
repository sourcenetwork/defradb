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

package test_explain_execute

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

func TestExecuteExplainRequestWithBothLimitAndOffsetOnParent(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			add2AddressDocuments(),
			add2AuthorContactDocuments(),
			add2AuthorDocuments(),
			add3BookDocuments(),

			&action.ExplainRequest{
				Request: `query @explain(type: execute) {
					Book(limit: 1, offset: 1) {
						name
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
									"limitNode": dataMap{
										"iterations": uint64(2),
										"selectNode": dataMap{
											"iterations":    uint64(2),
											"filterMatches": uint64(2),
											"scanNode": dataMap{
												"iterations":   uint64(2),
												"docFetches":   uint64(2),
												"fieldFetches": uint64(7),
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

func TestExecuteExplainRequestWithBothLimitAndOffsetOnParentAndLimitOnChild(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			add2AddressDocuments(),
			add2AuthorContactDocuments(),
			add2AuthorDocuments(),
			add3ArticleDocuments(),

			&action.ExplainRequest{
				Request: `query @explain(type: execute) {
					Author(limit: 1, offset: 1) {
						name
						articles(limit: 1) {
							name
						}
					}
				}`,

				ExpectedFullGraph: dataMap{
					"explain": dataMap{
						"executionSuccess": true,
						"planExecutions":   uint64(2),
						"sizeOfResult":     1,
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"limitNode": dataMap{
										"iterations": uint64(2),
										"selectNode": dataMap{
											"iterations":    uint64(2),
											"filterMatches": uint64(2),
											"typeIndexJoin": dataMap{
												"iterations": uint64(2),
												"typeJoinMany": dataMap{
													"root": dataMap{
														"scanNode": dataMap{
															"iterations":   uint64(2),
															"docFetches":   uint64(2),
															"fieldFetches": uint64(8),
															"indexFetches": uint64(0),
														},
													},
													"subType": dataMap{
														"selectTopNode": dataMap{
															"limitNode": dataMap{
																"iterations": uint64(4),
																"selectNode": dataMap{
																	"iterations":    uint64(2),
																	"filterMatches": uint64(2),
																	"scanNode": dataMap{
																		"iterations":   uint64(2),
																		"docFetches":   uint64(4),
																		"fieldFetches": uint64(12),
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
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

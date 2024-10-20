// Copyright 2024 Democratized Data Foundation
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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

func TestExecuteExplainRequest_WithMaxOfInlineArrayField_Succeeds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) request with max on an inline array.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			create2AddressDocuments(),
			create2AuthorContactDocuments(),
			create2AuthorDocuments(),
			create3BookDocuments(),

			testUtils.ExplainRequest{
				Request: `query @explain(type: execute) {
					Book {
						name
						MaxChapterPages: _max(chapterPages: {})
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
									"maxNode": dataMap{
										"iterations": uint64(4),
										"selectNode": dataMap{
											"iterations":    uint64(4),
											"filterMatches": uint64(3),
											"scanNode": dataMap{
												"iterations":   uint64(4),
												"docFetches":   uint64(3),
												"fieldFetches": uint64(5),
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

func TestExecuteExplainRequest_MaxOfRelatedOneToManyField_Succeeds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) request with max of a related one to many field.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,
			create2AddressDocuments(),
			create2AuthorContactDocuments(),
			create2AuthorDocuments(),
			create3ArticleDocuments(),

			testUtils.ExplainRequest{
				Request: `query @explain(type: execute) {
					Author {
						name
						MaxPages: _max(
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
									"maxNode": dataMap{
										"iterations": uint64(3),
										"selectNode": dataMap{
											"iterations":    uint64(3),
											"filterMatches": uint64(2),
											"typeIndexJoin": dataMap{
												"iterations": uint64(3),
												"scanNode": dataMap{
													"iterations":   uint64(3),
													"docFetches":   uint64(2),
													"fieldFetches": uint64(2),
													"indexFetches": uint64(0),
												},
												"subTypeScanNode": dataMap{
													"iterations":   uint64(5),
													"docFetches":   uint64(6),
													"fieldFetches": uint64(9),
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
	}

	explainUtils.ExecuteTestCase(t, test)
}

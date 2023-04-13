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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestExecuteExplainRequestWithSumOfInlineArrayField(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) request with sum on an inline array.",

		Actions: []any{
			gqlSchemaExecuteExplain(),

			// Books
			create3BookDocuments(),

			testUtils.Request{
				Request: `query @explain(type: execute) {
					Book {
						name
						NotSureWhySomeoneWouldSumTheChapterPagesButHereItIs: _sum(chapterPages: {})
					}
				}`,

				Results: []dataMap{
					{
						"explain": dataMap{
							"executionSuccess": true,
							"sizeOfResult":     3,
							"planExecutions":   uint64(4),
							"selectTopNode": dataMap{
								"sumNode": dataMap{
									"iterations": uint64(4),
									"selectNode": dataMap{
										"iterations":    uint64(4),
										"filterMatches": uint64(3),
										"scanNode": dataMap{
											"iterations":    uint64(4),
											"docFetches":    uint64(4),
											"filterMatches": uint64(3),
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

	executeTestCase(t, test)
}

func TestExecuteExplainRequestSumOfRelatedOneToManyField(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) request with sum of a related one to many field.",

		Actions: []any{
			gqlSchemaExecuteExplain(),

			// Articles
			create3ArticleDocuments(),

			// Authors
			create2AuthorDocuments(),

			testUtils.Request{
				Request: `query @explain(type: execute) {
					Author {
						name
						TotalPages: _sum(
							articles: {
								field: pages,
							}
						)
					}
				}`,

				Results: []dataMap{
					{
						"explain": dataMap{
							"executionSuccess": true,
							"sizeOfResult":     2,
							"planExecutions":   uint64(3),
							"selectTopNode": dataMap{
								"sumNode": dataMap{
									"iterations": uint64(3),
									"selectNode": dataMap{
										"iterations":    uint64(3),
										"filterMatches": uint64(2),
										"typeIndexJoin": dataMap{
											"iterations": uint64(3),
											"scanNode": dataMap{
												"iterations":    uint64(3),
												"docFetches":    uint64(3),
												"filterMatches": uint64(2),
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

	executeTestCase(t, test)
}

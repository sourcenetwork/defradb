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

func TestExecuteExplainRequestWithBothLimitAndOffsetOnParent(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) with both limit and offset on parent.",

		Actions: []any{
			gqlSchemaExecuteExplain(),

			// Books
			create3BookDocuments(),

			testUtils.Request{
				Request: `query @explain(type: execute) {
					Book(limit: 1, offset: 1) {
						name
					}
				}`,

				Results: []dataMap{
					{
						"explain": dataMap{
							"executionSuccess": true,
							"sizeOfResult":     1,
							"planExecutions":   uint64(2),
							"selectTopNode": dataMap{
								"limitNode": dataMap{
									"iterations": uint64(2),
									"selectNode": dataMap{
										"iterations":    uint64(2),
										"filterMatches": uint64(2),
										"scanNode": dataMap{
											"iterations":    uint64(2),
											"docFetches":    uint64(2),
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
	}

	executeTestCase(t, test)
}

func TestExecuteExplainRequestWithBothLimitAndOffsetOnParentAndLimitOnChild(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) with both limit and offset on parent and limit on child.",

		Actions: []any{
			gqlSchemaExecuteExplain(),

			// Articles
			create3ArticleDocuments(),

			// Authors
			create2AuthorDocuments(),

			testUtils.Request{
				Request: `query @explain(type: execute) {
					Author(limit: 1, offset: 1) {
						name
						articles(limit: 1) {
							name
						}
					}
				}`,

				Results: []dataMap{
					{
						"explain": dataMap{
							"executionSuccess": true,
							"planExecutions":   uint64(2),
							"sizeOfResult":     1,
							"selectTopNode": dataMap{
								"limitNode": dataMap{
									"iterations": uint64(2),
									"selectNode": dataMap{
										"iterations":    uint64(2),
										"filterMatches": uint64(2),
										"typeIndexJoin": dataMap{
											"iterations": uint64(2),
											"scanNode": dataMap{
												"iterations":    uint64(2),
												"docFetches":    uint64(2),
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

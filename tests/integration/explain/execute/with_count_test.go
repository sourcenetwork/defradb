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
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

func TestExecuteExplainRequestWithCountOnOneToManyRelation(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) request with count on one to many relation.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			create2AuthorDocuments(),
			create3BookDocuments(),

			testUtils.ExplainRequest{
				Request: `query @explain(type: execute) {
					Author {
						name
						numberOfBooks: _count(books: {})
					}
				}`,

				ExpectedFullGraph: dataMap{
					"explain": dataMap{
						"executionSuccess": true,
						"sizeOfResult":     2,
						"planExecutions":   uint64(3),
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"countNode": dataMap{
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
													"fieldFetches": uint64(6),
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

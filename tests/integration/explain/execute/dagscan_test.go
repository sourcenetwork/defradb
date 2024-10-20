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

func TestExecuteExplainCommitsDagScan(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) commits request - dagScan.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			create2AddressDocuments(),
			create2AuthorContactDocuments(),
			create2AuthorDocuments(),

			testUtils.ExplainRequest{
				Request: `query @explain(type: execute) {
					commits (docID: "bae-333455ca-1563-54c3-85a4-1db7ea4e9c59") {
						links {
							cid
						}
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
									"selectNode": dataMap{
										"iterations":    uint64(6),
										"filterMatches": uint64(5),
										"dagScanNode": dataMap{
											"iterations": uint64(6),
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

func TestExecuteExplainLatestCommitsDagScan(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) latest commits request - dagScan.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			create2AddressDocuments(),
			create2AuthorContactDocuments(),
			create2AuthorDocuments(),

			testUtils.ExplainRequest{
				Request: `query @explain(type: execute) {
					latestCommits(docID: "bae-333455ca-1563-54c3-85a4-1db7ea4e9c59") {
						cid
						links {
							cid
						}
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
									"selectNode": dataMap{
										"iterations":    uint64(2),
										"filterMatches": uint64(1),
										"dagScanNode": dataMap{
											"iterations": uint64(2),
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

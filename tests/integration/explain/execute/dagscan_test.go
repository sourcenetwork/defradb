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

			// Authors
			create2AuthorDocuments(),

			testUtils.ExplainRequest{
				Request: `query @explain(type: execute) {
					commits (dockey: "bae-7f54d9e0-cbde-5320-aa6c-5c8895a89138") {
						links {
							cid
						}
					}
				}`,

				ExpectedFullGraph: []dataMap{
					{
						"explain": dataMap{
							"executionSuccess": true,
							"sizeOfResult":     5,
							"planExecutions":   uint64(6),
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
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestExecuteExplainLatestCommitsDagScan(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) latest commits request - dagScan.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			// Author
			create2AuthorDocuments(),

			testUtils.ExplainRequest{
				Request: `query @explain(type: execute) {
					latestCommits(dockey: "bae-7f54d9e0-cbde-5320-aa6c-5c8895a89138") {
						cid
						links {
							cid
						}
					}
				}`,

				ExpectedFullGraph: []dataMap{
					{
						"explain": dataMap{
							"executionSuccess": true,
							"sizeOfResult":     1,
							"planExecutions":   uint64(2),
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
	}

	explainUtils.ExecuteTestCase(t, test)
}

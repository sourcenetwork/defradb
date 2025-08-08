// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain_debug

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

var dagScanPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"selectTopNode": dataMap{
					"selectNode": dataMap{
						"dagScanNode": dataMap{},
					},
				},
			},
		},
	},
}

func TestDebugExplainCommitsDagScanQueryOp(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) commits query-op.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					commits (docID: "bae-9e70648f-c722-5875-97f5-574ec6f703e9", fieldName: "name") {
						links {
							cid
						}
					}
				}`,

				ExpectedFullGraph: dagScanPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainCommitsDagScanQueryOpWithoutField(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) commits query-op with only docID (no field).",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					commits (docID: "bae-9e70648f-c722-5875-97f5-574ec6f703e9") {
						links {
							cid
						}
					}
				}`,

				ExpectedFullGraph: dagScanPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainLatestCommitsDagScanQueryOp(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) latestCommits query-op.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					latestCommits(docID: "bae-9e70648f-c722-5875-97f5-574ec6f703e9", fieldName: "name") {
						cid
						links {
							cid
						}
					}
				}`,

				ExpectedFullGraph: dagScanPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainLatestCommitsDagScanQueryOpWithoutField(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) latestCommits query-op with only docID (no field).",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					latestCommits(docID: "bae-9e70648f-c722-5875-97f5-574ec6f703e9") {
						cid
						links {
							cid
						}
					}
				}`,

				ExpectedFullGraph: dagScanPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainLatestCommitsDagScanWithoutDocID_Failure(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) latestCommits query without docID.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					latestCommits(fieldName: "name") {
						cid
						links {
							cid
						}
					}
				}`,

				ExpectedError: "Field \"latestCommits\" argument \"docID\" of type \"ID!\" is required but not provided.",
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainLatestCommitsDagScanWithoutAnyArguments_Failure(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) latestCommits query without any arguments.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					latestCommits {
						cid
						links {
							cid
						}
					}
				}`,

				ExpectedError: "Field \"latestCommits\" argument \"docID\" of type \"ID!\" is required but not provided.",
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

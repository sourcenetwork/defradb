// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain_default

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

func TestDefaultExplainCommitsDagScanQueryOp(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) commits query-op.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					commits (docID: "bae-9e70648f-c722-5875-97f5-574ec6f703e9", fieldName: "name") {
						links {
							cid
						}
					}
				}`,

				ExpectedPatterns: dagScanPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "dagScanNode",
						IncludeChildNodes: true, // Shouldn't have any as this is the last node in the chain.
						ExpectedAttributes: dataMap{
							"cid":       nil,
							"fieldName": "name",
							"prefixes": []string{
								"/d/bae-9e70648f-c722-5875-97f5-574ec6f703e9",
							},
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainCommitsDagScanQueryOpWithoutField(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) commits query-op with only docID (no field).",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					commits (docID: "bae-9e70648f-c722-5875-97f5-574ec6f703e9") {
						links {
							cid
						}
					}
				}`,

				ExpectedPatterns: dagScanPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "dagScanNode",
						IncludeChildNodes: true, // Shouldn't have any as this is the last node in the chain.
						ExpectedAttributes: dataMap{
							"cid":       nil,
							"fieldName": nil,
							"prefixes": []string{
								"/d/bae-9e70648f-c722-5875-97f5-574ec6f703e9",
							},
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainLatestCommitsDagScanQueryOp(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) latestCommits query-op.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					latestCommits(docID: "bae-9e70648f-c722-5875-97f5-574ec6f703e9", fieldName: "name") {
						cid
						links {
							cid
						}
					}
				}`,

				ExpectedPatterns: dagScanPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "dagScanNode",
						IncludeChildNodes: true, // Shouldn't have any as this is the last node in the chain.
						ExpectedAttributes: dataMap{
							"cid":       nil,
							"fieldName": "name",
							"prefixes": []string{
								"/d/bae-9e70648f-c722-5875-97f5-574ec6f703e9",
							},
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainLatestCommitsDagScanQueryOpWithoutField(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) latestCommits query-op with only docID (no field).",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					latestCommits(docID: "bae-9e70648f-c722-5875-97f5-574ec6f703e9") {
						cid
						links {
							cid
						}
					}
				}`,

				ExpectedPatterns: dagScanPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "dagScanNode",
						IncludeChildNodes: true, // Shouldn't have any as this is the last node in the chain.
						ExpectedAttributes: dataMap{
							"cid":       nil,
							"fieldName": "_C",
							"prefixes": []string{
								"/d/bae-9e70648f-c722-5875-97f5-574ec6f703e9/C",
							},
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainLatestCommitsDagScanWithoutDocID_Failure(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) latestCommits query without docID.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
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

func TestDefaultExplainLatestCommitsDagScanWithoutAnyArguments_Failure(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) latestCommits query without any arguments.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
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

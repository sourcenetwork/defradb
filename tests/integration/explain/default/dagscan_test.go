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

package test_explain_default

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					_commits (
						docID: "bae-9e70648f-c722-5875-97f5-574ec6f703e9", 
						filter: {fieldName: {_eq: "name"}}
					) {
						links {
							cid
						}
					}
				}`,

				ExpectedPatterns: dagScanPattern,

				ExpectedTargets: []action.PlanNodeTargetCase{
					{
						TargetNodeName:    "dagScanNode",
						IncludeChildNodes: true, // Shouldn't have any as this is the last node in the chain.
						ExpectedAttributes: dataMap{
							"cid": nil,
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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					_commits (docID: "bae-9e70648f-c722-5875-97f5-574ec6f703e9") {
						links {
							cid
						}
					}
				}`,

				ExpectedPatterns: dagScanPattern,

				ExpectedTargets: []action.PlanNodeTargetCase{
					{
						TargetNodeName:    "dagScanNode",
						IncludeChildNodes: true, // Shouldn't have any as this is the last node in the chain.
						ExpectedAttributes: dataMap{
							"cid": nil,
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

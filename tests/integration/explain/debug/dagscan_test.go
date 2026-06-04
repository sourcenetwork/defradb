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

package test_explain_debug

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

func TestDebugExplainCommitsDagScanQueryOp(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					_commits (docID: "bae-9e70648f-c722-5875-97f5-574ec6f703e9", filter: {fieldName: {_eq: "name"}}) {
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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					_commits (docID: "bae-9e70648f-c722-5875-97f5-574ec6f703e9") {
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

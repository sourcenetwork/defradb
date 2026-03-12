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

var upsertPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"upsertNode": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"scanNode": dataMap{},
						},
					},
				},
			},
		},
	},
}

func TestDebugExplainMutationRequest_WithUpsert_Succeeds(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `mutation @explain(type: debug) {
					upsert_Author(
						filter: {name: {_eq: "Bob"}},
						update: {age: 59},
						add: {name: "Bob", age: 59}
					) {
						_docID
						name
						age
					}
				}`,

				ExpectedPatterns: upsertPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

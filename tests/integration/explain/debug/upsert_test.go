// Copyright 2024 Democratized Data Foundation
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

		Description: "Explain (debug) mutation request with upsert.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain(type: debug) {
					upsert_Author(
						filter: {name: {_eq: "Bob"}},
						update: {age: 59},
						create: {name: "Bob", age: 59}
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

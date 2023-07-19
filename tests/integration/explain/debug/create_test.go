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

var createPattern = dataMap{
	"explain": dataMap{
		"createNode": dataMap{
			"selectTopNode": dataMap{
				"selectNode": dataMap{
					"scanNode": dataMap{},
				},
			},
		},
	},
}

func TestDebugExplainMutationRequestWithCreate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Explain (debug) mutation request with create.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain(type: debug) {
					create_Author(data: "{\"name\": \"Shahzad Lone\",\"age\": 27,\"verified\": true}") {
						name
						age
					}
				}`,

				ExpectedPatterns: []dataMap{createPattern},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainMutationRequestDoesNotCreateDocGivenDuplicate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Explain (debug) mutation request with create, document exists.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain(type: debug) {
					create_Author(data: "{\"name\": \"Shahzad Lone\",\"age\": 27}") {
						name
						age
					}
				}`,

				ExpectedPatterns: []dataMap{createPattern},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

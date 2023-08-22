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

func TestExecuteExplainMutationRequestWithCreate(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) mutation request with create.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{
				Request: `mutation @explain(type: execute) {
					create_Author(data: "{\"name\": \"Shahzad Lone\",\"age\": 27,\"verified\": true}") {
						name
					}
				}`,

				ExpectedFullGraph: []dataMap{
					{
						"explain": dataMap{
							"executionSuccess": true,
							"sizeOfResult":     1,
							"planExecutions":   uint64(2),
							"createNode": dataMap{
								"iterations": uint64(2),
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"iterations":    uint64(1),
										"filterMatches": uint64(1),
										"scanNode": dataMap{
											"iterations":   uint64(1),
											"docFetches":   uint64(1),
											"fieldFetches": uint64(1),
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
	}

	explainUtils.ExecuteTestCase(t, test)
}

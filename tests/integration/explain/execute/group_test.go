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

func TestExecuteExplainRequestWithGroup(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) request with groupBy.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			// Books
			create2AddressDocuments(),

			testUtils.ExplainRequest{
				Request: `query @explain(type: execute) {
					ContactAddress(groupBy: [country]) {
						country
						_group {
							city
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
									"groupNode": dataMap{
										"iterations":            uint64(2),
										"groups":                uint64(1),
										"childSelections":       uint64(1),
										"hiddenBeforeOffset":    uint64(0),
										"hiddenAfterLimit":      uint64(0),
										"hiddenChildSelections": uint64(0),
										"selectNode": dataMap{
											"iterations":    uint64(3),
											"filterMatches": uint64(2),
											"scanNode": dataMap{
												"iterations":   uint64(4),
												"docFetches":   uint64(2),
												"fieldFetches": uint64(4),
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
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

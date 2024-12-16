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

func TestExecuteExplainMutationRequestWithUpdateUsingIDs(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) mutation request with update using document IDs.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			// Addresses
			create2AddressDocuments(),

			testUtils.ExplainRequest{
				Request: `mutation @explain(type: execute) {
					update_ContactAddress(
						docID: [
							"bae-14f20db7-3654-58de-9156-596ef2cfd790",
							"bae-49f715e7-7f01-5509-a213-ed98cb81583f"
						],
						input: {country: "USA"}
					) {
						country
						city
					}
				}`,

				ExpectedFullGraph: dataMap{
					"explain": dataMap{
						"executionSuccess": true,
						"sizeOfResult":     1,
						"planExecutions":   uint64(2),
						"operationNode": []dataMap{
							{
								"updateNode": dataMap{
									"iterations": uint64(3),
									"updates":    uint64(2),
									"selectTopNode": dataMap{
										"selectNode": dataMap{
											"iterations":    uint64(6),
											"filterMatches": uint64(4),
											"scanNode": dataMap{
												"iterations":   uint64(6),
												"docFetches":   uint64(4),
												"fieldFetches": uint64(8),
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

func TestExecuteExplainMutationRequestWithUpdateUsingFilter(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) mutation request with update using filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			// Addresses
			create2AddressDocuments(),

			testUtils.ExplainRequest{
				Request: `mutation @explain(type: execute) {
					update_ContactAddress(
						filter: {
							city: {
								_eq: "Waterloo"
							}
						},
						input: {country: "USA"}
					) {
						country
						city
					}
				}`,

				ExpectedFullGraph: dataMap{
					"explain": dataMap{
						"executionSuccess": true,
						"sizeOfResult":     1,
						"planExecutions":   uint64(2),
						"operationNode": []dataMap{
							{
								"updateNode": dataMap{
									"iterations": uint64(2),
									"updates":    uint64(1),
									"selectTopNode": dataMap{
										"selectNode": dataMap{
											"iterations":    uint64(4),
											"filterMatches": uint64(2),
											"scanNode": dataMap{
												"iterations":   uint64(4),
												"docFetches":   uint64(4),
												"fieldFetches": uint64(8),
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

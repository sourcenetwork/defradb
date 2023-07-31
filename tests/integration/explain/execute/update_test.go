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

		Description: "Explain (execute) mutation request with update using ids.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			// Addresses
			create2AddressDocuments(),

			testUtils.ExplainRequest{
				Request: `mutation @explain(type: execute) {
					update_ContactAddress(
						ids: [
							"bae-c8448e47-6cd1-571f-90bd-364acb80da7b",
							"bae-f01bf83f-1507-5fb5-a6a3-09ecffa3c692"
						],
						data: "{\"country\": \"USA\"}"
					) {
						country
						city
					}
				}`,

				ExpectedFullGraph: []dataMap{
					{
						"explain": dataMap{
							"executionSuccess": true,
							"sizeOfResult":     2,
							"planExecutions":   uint64(3),
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
						data: "{\"country\": \"USA\"}"
					) {
						country
						city
					}
				}`,

				ExpectedFullGraph: []dataMap{
					{
						"explain": dataMap{
							"executionSuccess": true,
							"sizeOfResult":     1,
							"planExecutions":   uint64(2),
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
											"fieldFetches": uint64(6),
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

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

package test_explain_execute

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

func TestExecuteExplainMutationRequestWithUpdateUsingIDs(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			// Addresses
			add2AddressDocuments(),

			&action.ExplainRequest{
				Request: `mutation @explain(type: execute) {
					update_ContactAddress(
						docID: [
							"bae-186c2484-c3ea-5993-95d6-cb886e1b13a1",
							"bae-78bc4454-19a6-58ed-9e18-f0ca175dd12c"
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
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"iterations":    uint64(3),
										"filterMatches": uint64(2),
										"updateNode": dataMap{
											"iterations": uint64(3),
											"updates":    uint64(2),
											"selectTopNode": dataMap{
												"selectNode": dataMap{
													"iterations":    uint64(3),
													"filterMatches": uint64(2),
													"scanNode": dataMap{
														"iterations":   uint64(3),
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
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestExecuteExplainMutationRequestWithUpdateUsingFilter(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			// Addresses
			add2AddressDocuments(),

			&action.ExplainRequest{
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
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"iterations":    uint64(2),
										"filterMatches": uint64(1),
										"updateNode": dataMap{
											"iterations": uint64(2),
											"updates":    uint64(1),
											"selectTopNode": dataMap{
												"selectNode": dataMap{
													"iterations":    uint64(2),
													"filterMatches": uint64(1),
													"scanNode": dataMap{
														"iterations":   uint64(2),
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
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

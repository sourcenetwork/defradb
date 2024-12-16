// Copyright 2024 Democratized Data Foundation
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

func TestExecuteExplainMutationRequest_WithUpsertAndMatchingFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) mutation request with upsert and matching filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			// Addresses
			create2AddressDocuments(),

			testUtils.ExplainRequest{
				Request: `mutation @explain(type: execute) {
					upsert_ContactAddress(
						filter: {city: {_eq: "Waterloo"}},
						create: {city: "Waterloo", country: "USA"},
						update: {country: "USA"}
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
								"upsertNode": dataMap{
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

func TestExecuteExplainMutationRequest_WithUpsertAndNoMatchingFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) mutation request with upsert and no matching filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{
				Request: `mutation @explain(type: execute) {
					upsert_ContactAddress(
						filter: {city: {_eq: "Waterloo"}},
						create: {city: "Waterloo", country: "USA"},
						update: {country: "USA"}
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
								"upsertNode": dataMap{
									"selectTopNode": dataMap{
										"selectNode": dataMap{
											"iterations":    uint64(3),
											"filterMatches": uint64(1),
											"scanNode": dataMap{
												"iterations":   uint64(3),
												"docFetches":   uint64(1),
												"fieldFetches": uint64(2),
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

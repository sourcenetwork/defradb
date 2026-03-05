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

func TestExecuteExplainMutationRequest_WithUpsertAndMatchingFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			// Addresses
			add2AddressDocuments(),

			&action.ExplainRequest{
				Request: `mutation @explain(type: execute) {
					upsert_ContactAddress(
						filter: {city: {_eq: "Waterloo"}},
						add: {city: "Waterloo", country: "USA"},
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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{
				Request: `mutation @explain(type: execute) {
					upsert_ContactAddress(
						filter: {city: {_eq: "Waterloo"}},
						add: {city: "Waterloo", country: "USA"},
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

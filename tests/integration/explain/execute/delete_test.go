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

func TestExecuteExplainMutationRequestWithDeleteUsingID(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			// Addresses
			add2AddressDocuments(),

			&action.ExplainRequest{
				Request: `mutation @explain(type: execute) {
					delete_ContactAddress(docID: ["bae-78bc4454-19a6-58ed-9e18-f0ca175dd12c"]) {
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
								"deleteNode": dataMap{
									"iterations": uint64(2),
									"selectTopNode": dataMap{
										"selectNode": dataMap{
											"iterations":    uint64(2),
											"filterMatches": uint64(1),
											"scanNode": dataMap{
												"iterations":   uint64(2),
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

func TestExecuteExplainMutationRequestWithDeleteUsingFilter(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			add2AddressDocuments(),
			add2AuthorContactDocuments(),
			add2AuthorDocuments(),

			&action.ExplainRequest{
				Request: `mutation @explain(type: execute) {
					delete_Author(filter: {name: {_like: "%Funke%"}}) {
						name
					}
				}`,

				ExpectedFullGraph: dataMap{
					"explain": dataMap{
						"executionSuccess": true,
						"sizeOfResult":     1,
						"planExecutions":   uint64(2),
						"operationNode": []dataMap{
							{
								"deleteNode": dataMap{
									"iterations": uint64(2),
									"selectTopNode": dataMap{
										"selectNode": dataMap{
											"iterations":    uint64(2),
											"filterMatches": uint64(1),
											"scanNode": dataMap{
												"iterations":   uint64(2),
												"docFetches":   uint64(2),
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

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

func TestExecuteExplainQueryDeletedDocs(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,
			add2AddressDocuments(),
			&action.Request{
				Request: `mutation  {
					delete_ContactAddress(docID: ["bae-78bc4454-19a6-58ed-9e18-f0ca175dd12c"]) {
						_docID
					}
				}`,
				Results: map[string]any{
					"delete_ContactAddress": []map[string]any{
						{"_docID": "bae-78bc4454-19a6-58ed-9e18-f0ca175dd12c"},
					},
				},
			},
			&action.ExplainRequest{
				Request: `query @explain(type: execute) {
					ContactAddress(showDeleted: true) {
						city
						country
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
	}

	explainUtils.ExecuteTestCase(t, test)
}

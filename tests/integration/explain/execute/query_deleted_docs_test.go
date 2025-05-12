// Copyright 2023 Democratized Data Foundation
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

func TestExecuteExplainQueryDeletedDocs(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Explain (execute) query with deleted documents.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,
			create2AddressDocuments(),
			testUtils.Request{
				Request: `mutation  {
					delete_ContactAddress(docID: ["bae-0d4eb5f3-499d-553d-9fb9-80a19463ec9a"]) {
						_docID
					}
				}`,
				Results: map[string]any{
					"delete_ContactAddress": []map[string]any{
						{"_docID": "bae-0d4eb5f3-499d-553d-9fb9-80a19463ec9a"},
					},
				},
			},
			testUtils.ExplainRequest{
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

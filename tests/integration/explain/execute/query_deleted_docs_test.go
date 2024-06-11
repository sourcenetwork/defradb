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
					delete_ContactAddress(docIDs: ["bae-49f715e7-7f01-5509-a213-ed98cb81583f"]) {
						_docID
					}
				}`,
				Results: []map[string]any{
					{"_docID": "bae-49f715e7-7f01-5509-a213-ed98cb81583f"},
				},
			},
			testUtils.ExplainRequest{
				Request: `query @explain(type: execute) {
					ContactAddress(showDeleted: true) {
						city
						country
					}
				}`,

				ExpectedFullGraph: []dataMap{
					{
						"explain": dataMap{
							"executionSuccess": true,
							"sizeOfResult":     2,
							"planExecutions":   uint64(3),
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
	}

	explainUtils.ExecuteTestCase(t, test)
}

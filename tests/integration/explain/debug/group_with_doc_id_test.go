// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain_debug

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

func TestDebugExplainRequestWithDocIDOnParentGroupBy(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with a document ID on parent groupBy.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(
						groupBy: [age],
						docID: "bae-6a4c5bc5-b044-5a03-a868-8260af6f2254"
					) {
						age
						_group {
							name
						}
					}
				}`,

				ExpectedPatterns: []dataMap{groupPattern},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithDocIDsAndFilterOnParentGroupBy(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with document IDs and filter on parent groupBy.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(
						groupBy: [age],
						filter: {age: {_eq: 20}},
						docIDs: [
							"bae-6a4c5bc5-b044-5a03-a868-8260af6f2254",
							"bae-4ea9d148-13f3-5a48-a0ef-9ffd344caeed"
						]
					) {
						age
						_group {
							name
						}
					}
				}`,

				ExpectedPatterns: []dataMap{groupPattern},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

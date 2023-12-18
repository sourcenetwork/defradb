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

func TestDebugExplainRequestWithDockeysOnInnerGroupSelection(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with docIDs on inner _group.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(
						groupBy: [age]
					) {
						age
						_group(docIDs: ["bae-6a4c5bc5-b044-5a03-a868-8260af6f2254"]) {
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

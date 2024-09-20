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

func TestDebugExplainMutationRequestWithDeleteHavingNoSubSelection(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) multation request with delete having no sub-selection.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain(type: debug) {
					delete_Author(
						docID: [
							"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
							"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f"
						]
					)
				}`,

				ExpectedError: "Field \"delete_Author\" of type \"[Author]\" must have a sub selection.",
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

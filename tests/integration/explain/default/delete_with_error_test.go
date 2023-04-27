// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain_default

import (
	"testing"

	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

func TestDefaultExplainMutationRequestWithDeleteHavingNoSubSelection(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) multation request with delete having no sub-selection.",

		Request: `mutation @explain {
				delete_author(ids: ["bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d", "bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f"])
			}`,

		Docs: map[int][]string{
			2: {
				// bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f
				`{
						"name": "Lone",
						"age":  26,
						"verified": false
					}`,
				// "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
				`{
						"name":     "Shahzad Lone",
						"age":      27,
						"verified": true
					}`,
			},
		},

		ExpectedError: "Field \"delete_author\" of type \"[author]\" must have a sub selection.",
	}

	runExplainTest(t, test)
}

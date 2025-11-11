// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package branchables

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsBranchables_WithCidAndDocIDParam(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users @branchable {
						name: String
						age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
			},
			testUtils.Request{
				// This request uses the document's docID, and the collection's cid.
				// It would be very nice if this worked:
				// https://github.com/sourcenetwork/defradb/issues/3213
				Request: `query {
						_commits(
							docID: "bae-f895da58-3326-510a-87f3-d043ff5424ea",
							cid: "bafyreigvb5smrzfjpunohvfsoi3czzjy4dl3cyhzkutth4lx27wnasnqxa"
						) {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
				ExpectedError: "cid either does not exist or belong to document",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

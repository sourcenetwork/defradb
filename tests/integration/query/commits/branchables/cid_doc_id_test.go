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

package branchables

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsBranchables_WithCidAndDocIDParam(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users @branchable {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
			},
			&action.Request{
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

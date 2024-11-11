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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsBranchables_WithCidAndDocIDParam(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
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
						commits(
							docID: "bae-0b2f15e5-bfe7-5cb7-8045-471318d7dbc3",
							cid: "bafyreifi7borlnkazxrcohgl7r36cm5ga7moyiiajov3om7urexbx7cyl4"
						) {
							cid
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

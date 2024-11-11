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

func TestQueryCommitsBranchables_WithCidParam(t *testing.T) {
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
				Request: `query {
						commits(
							cid: "bafyreifi7borlnkazxrcohgl7r36cm5ga7moyiiajov3om7urexbx7cyl4"
						) {
							cid
							collectionID
							docID
							fieldName
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": testUtils.NewUniqueCid("collection"),
							// Extra params are used to verify this is a collection level cid
							"collectionID": int64(1),
							"docID":        nil,
							"fieldName":    nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

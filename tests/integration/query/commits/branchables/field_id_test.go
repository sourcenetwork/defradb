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

func TestQueryCommitsBranchables_WithFieldName(t *testing.T) {
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
							fieldName: null
						) {
							schemaVersionId
							docID
							fieldName
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							// Extra params are used to verify this is a collection level cid
							"schemaVersionId": "bafkreihhd6bqrjhl5zidwztgxzeseveplv3cj3fwtn3unjkdx7j2vr2vrq",
							"docID":           nil,
							"fieldName":       nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

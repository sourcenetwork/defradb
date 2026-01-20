// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithCidOfBranchableCollectionAndDocID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users @branchable {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"name": "Freddddd"
				}`,
			},
			testUtils.Request{
				// This is the cid of the collection-commit when the second doc (John) is created.
				// Without the docID param both John and Fred should be returned.
				Request: `query {
					Users (
							cid: "bafyreicxlkbypv4hjc2trunobsugx63no47322xfjjazj3v5y6hqeck3d4",
							docID: "bae-235c64e3-abf7-549c-9aff-971c8afdfa3f"
						) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithCidOfBranchableCollectionAndDocID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
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
							cid: "bafyreiboen2mw2unu4fty2pyyd5nicqi57vcdahrrag6bjm54md5myj54u",
							docID: "bae-3a7df128-bfa9-559a-a9c5-96f2bf6d1038"
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

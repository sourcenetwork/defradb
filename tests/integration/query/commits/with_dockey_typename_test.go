// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package commits

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsWithDockeyWithTypeName(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey and typename",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
				Request: `query {
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
							cid
							__typename
						}
					}`,
				Results: []map[string]any{
					{
						"cid":        "bafybeic5oodfpnixl6uf4bi63m3eouuhj3gafudlsd4tqryhx2wy7rczoe",
						"__typename": "Commit",
					},
					{
						"cid":        "bafybeifukwb3t73k7pph3ctp5khosoycp53ywjl6btravzk6decggkjtl4",
						"__typename": "Commit",
					},
					{
						"cid":        "bafybeig3wrpwi6q7vjchizcwnenslasyxop6wey7jahbiszlubdglfq2fq",
						"__typename": "Commit",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

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

func TestQueryCommitsWithDocIDAndLinkCount(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple latest commits query with docID and link count",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
							cid
							_count(field: links)
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeiagew5ervkxkyz6sezgkrcrhv6towawpbirg6ziouah7lj5eg7lwy",
						"_count": 0,
					},
					{
						"cid":    "bafybeibsfjpv57libozcmiihs2ixgsoseu752ni6nkeqqv2nhgxgpe7l5m",
						"_count": 0,
					},
					{
						"cid":    "bafybeiaitgizimh6ttm3stfabpwqvg5tvty3j46p6hqbat4w7ippcbw4ky",
						"_count": 2,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

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

func TestQueryCommitsWithDockeyAndLimit(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey and limit",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"John",
						"Age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"Age":	22
				}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"Age":	23
				}`,
			},
			testUtils.Request{
				Request: ` {
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", limit: 2) {
							cid
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeihvifwxwmfuyeupebhkalie5odlafnzdjpmwqz5kuo5zld63ishve",
					},
					{
						"cid": "bafybeigtudzrntyslfdtukiobzen76iuhkmyo3x4eutx3igwzwhjekp2oi",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

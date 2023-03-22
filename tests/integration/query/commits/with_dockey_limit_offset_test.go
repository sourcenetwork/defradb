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

func TestQueryCommitsWithDockeyAndLimitAndOffset(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, limit and offset",
		Actions: []any{
			updateUserCollectionSchema(),
			createJohnDoc(),
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
						"Age": 22
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
						"Age": 23
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
						"Age": 24
					}`,
			},
			testUtils.Request{
				Request: ` {
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", limit: 2, offset: 1) {
							cid
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeifaxl4u5wmokgr4jviru6dz7teg7f2fomusxrvh7o5nh2a32jk3va",
					},
					{
						"cid": "bafybeihxc6ittcok3rnetguamxfzd3wa534z7zwqsaoppvawu7jx4rdy5u",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

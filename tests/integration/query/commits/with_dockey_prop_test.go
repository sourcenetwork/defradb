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

func TestQueryCommitsWithDockeyProperty(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple commits query with dockey property",
		Actions: []any{
			updateUserCollectionSchema(),
			createDoc("John", 21),
			testUtils.Request{
				Request: `query {
						commits {
							cid
							dockey
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeiaeic6vhiiw5zu6ju7e47cclvctn6t5pb36fj3mczchyhmctbrr6m",
						"dockey": "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
					},
					{
						"cid":    "bafybeibsaubd2ptp6qqsszv24p73j474amc4pll4oyssnpilofrl575hmy",
						"dockey": "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
					},
					{
						"cid":    "bafybeidcatznm2mlsymcytrh5fkpdrazensg5fsvn2uavcgiq2bf26lzey",
						"dockey": "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

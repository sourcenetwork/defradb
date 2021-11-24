// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package latest_commits

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/db/tests"
)

func TestQueryLatestCommits(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple latest commits query",
		Query: `query {
					latestCommits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
						cid
						links {
							cid
							name
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 21
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"cid": "bafkreiercmxn6e3qryxvuped5pplg733c5fj6gjypj5wykk63ouvcfb25m",
				"links": []map[string]interface{}{
					{
						"cid":  "bafybeiasnjaz6bohhhqopk77ksivqed5wgbog7575wunleaq57nar6otui",
						"name": "Age",
					},
					{
						"cid":  "bafybeifxin4fbdnc4hrn5tyimnzy53jj6oxtu5kpgohzv5y5wsrpjoih6a",
						"name": "Name",
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

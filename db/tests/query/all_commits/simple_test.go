// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package all_commits

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/db/tests"
)

func TestQueryAllCommitsSingleDAG(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple latest commits query",
		Query: `query {
					allCommits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
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
				"cid": "bafkreie33v5gnk4vlkzedrm5tofkhhwwrb3mlnpibfsa6qzgvhpiv5qzl4",
				"links": []map[string]interface{}{
					{
						"cid":  "bafybeiftyjqxyzqtfpi65kde4hla4xm3v4dvtr7fr2p2p5ng5lfg7rrcve",
						"name": "Age",
					},
					{
						"cid":  "bafybeierejzn3m6pesium3cml4flyjoe2wd2pxbmxxi5v42yqw2w4fpcxm",
						"name": "Name",
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryAllCommitsMultipleDAG(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple latest commits query",
		Query: `query {
					allCommits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
						cid
						height
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 21
			}`)},
		},
		Updates: map[int][]string{
			0: {
				// update to change age to 22 on document 0
				(`{"Age": 22}`),
			},
		},
		Results: []map[string]interface{}{
			{
				"cid":    "bafkreiap6dmsuhgzfbnj6zeobmhhmra37v3hokal7ef7dhaqp4qtpou2sa",
				"height": int64(2),
			},
			{
				"cid":    "bafkreie33v5gnk4vlkzedrm5tofkhhwwrb3mlnpibfsa6qzgvhpiv5qzl4",
				"height": int64(1),
			},
		},
	}

	executeTestCase(t, test)
}

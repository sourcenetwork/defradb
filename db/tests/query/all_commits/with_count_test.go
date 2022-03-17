// Copyright 2022 Democratized Data Foundation
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

func TestQueryAllCommitsSingleDAGWithLinkCount(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple latest commits query",
		Query: `query {
					allCommits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
						cid
						_count(field: links)
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
				"cid":    "bafybeih2egliqqrwwykitohimsvqgtg4dvx5ts5vraadhjdz5ls2cnvpnq",
				"_count": 2,
			},
		},
	}

	executeTestCase(t, test)
}

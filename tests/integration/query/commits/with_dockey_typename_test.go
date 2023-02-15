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
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with dockey and typename",
		Request: `query {
					commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
						cid
						__typename
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
			},
		},
		Results: []map[string]any{
			{
				"cid":        "bafybeicovjpmtwu544e7hzgg7mcwabstmugesi3n62ju6kbimcsjqp23gu",
				"__typename": "Commit",
			},
			{
				"cid":        "bafybeietvbhkavrhb6usprlsehh5cojgznzqv4zdah2bhbrmgc2ph3rxka",
				"__typename": "Commit",
			},
			{
				"cid":        "bafybeignirnk6wuxtg2fzwfbvs26wmrldlhpqj243kwnxb4ewafbae23m4",
				"__typename": "Commit",
			},
		},
	}

	executeTestCase(t, test)
}

// Copyright 2022 Democratized Data Foundation
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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryLastCommitsWithDockeyProperty(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple latest commits query with dockey property",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"John",
						"Age":	21
					}`,
			},
			testUtils.Request{
				Request: `query {
						latestCommits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
							dockey
						}
					}`,
				Results: []map[string]any{
					{
						"dockey": "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

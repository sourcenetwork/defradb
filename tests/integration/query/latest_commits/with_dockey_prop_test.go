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
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
				Request: `query {
						latestCommits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
							dockey
						}
					}`,
				Results: []map[string]any{
					{
						"dockey": "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

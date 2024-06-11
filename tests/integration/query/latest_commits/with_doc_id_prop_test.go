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

func TestQueryLastCommitsWithDocIDProperty(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple latest commits query with docID property",
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
						latestCommits(docID: "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3") {
							docID
						}
					}`,
				Results: []map[string]any{
					{
						"docID": "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

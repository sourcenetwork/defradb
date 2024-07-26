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

func TestQueryLastCommitsWithCollectionIdProperty(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple latest commits query with collectionID property",
		Actions: []any{
			updateUserCollectionSchema(),
			updateCompaniesCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
						"name":	"Source"
					}`,
			},
			testUtils.Request{
				Request: `query {
						latestCommits(docID: "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3") {
							collectionID
						}
					}`,
				Results: map[string]any{
					"latestCommits": []map[string]any{
						{
							"collectionID": int64(1),
						},
					},
				},
			},
			testUtils.Request{
				Request: `query {
						latestCommits(docID: "bae-f824cbf5-cc66-5e44-a84f-e71f72ff9841") {
							collectionID
						}
					}`,
				Results: map[string]any{
					"latestCommits": []map[string]any{
						{
							"collectionID": int64(2),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

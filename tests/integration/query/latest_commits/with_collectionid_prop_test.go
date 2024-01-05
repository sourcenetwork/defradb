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
						latestCommits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
							collectionID
						}
					}`,
				Results: []map[string]any{
					{
						"collectionID": int64(1),
					},
				},
			},
			testUtils.Request{
				Request: `query {
						latestCommits(docID: "bae-de8c99bf-ee0e-5655-8a72-919c2d459a30") {
							collectionID
						}
					}`,
				Results: []map[string]any{
					{
						"collectionID": int64(2),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

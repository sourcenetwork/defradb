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

func TestQueryCommitsWithDocIDWithTypeName(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID and typename",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
							cid
							__typename
						}
					}`,
				Results: []map[string]any{
					{
						"cid":        "bafybeibcr5lkdvcvtr67rpsnvn57hgrhlg36cnmmf7kywjekjodwxytpi4",
						"__typename": "Commit",
					},
					{
						"cid":        "bafybeifw7cu7uweruypv44on2zupjzolyqvyh4ookoeybkztzys67m4hwi",
						"__typename": "Commit",
					},
					{
						"cid":        "bafybeige7qoom3bgjitfisxvhbifou6n4tgguan3ihwbkz5mvbumndeiaa",
						"__typename": "Commit",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

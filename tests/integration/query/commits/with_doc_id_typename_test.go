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
						commits(docID: "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3") {
							cid
							__typename
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid":        "bafyreifzyy7bmpx2eywj4lznxzrzrvh6vrz6l7bhthkpexdq3wtho3vz6i",
							"__typename": "Commit",
						},
						{
							"cid":        "bafyreic2sba5sffkfnt32wfeoaw4qsqozjb5acwwtouxuzllb3aymjwute",
							"__typename": "Commit",
						},
						{
							"cid":        "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q",
							"__typename": "Commit",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

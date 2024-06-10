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
				Results: []map[string]any{
					{
						"cid":        "bafyreietqxguz3xlady4gfaqnbeamwsnrwfkufykkpprxej7a77ba7siay",
						"__typename": "Commit",
					},
					{
						"cid":        "bafyreidksmyoo6txzmcygby6quhdkzymlqoaxpg75ehlxjdneotjzbih6y",
						"__typename": "Commit",
					},
					{
						"cid":        "bafyreichxrfyhajs7rzp3wh5f2zrmt3zkjqan5dmxoy4qno5ozy7omzfpq",
						"__typename": "Commit",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

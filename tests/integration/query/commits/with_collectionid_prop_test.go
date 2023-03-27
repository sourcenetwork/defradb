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

func TestQueryCommitsWithCollectionID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple commits query with dockey property",
		Actions: []any{
			updateUserCollectionSchema(),
			updateCompaniesCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"John",
						"Age":	21
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
						"Name":	"Source"
					}`,
			},
			testUtils.Request{
				Request: `query {
						commits {
							collectionId
						}
					}`,
				Results: []map[string]any{
					{
						"collectionId": 1,
					},
					{
						"collectionId": 1,
					},
					{
						"collectionId": 1,
					},
					{
						"collectionId": 2,
					},
					{
						"collectionId": 2,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users", "companies"}, test)
}

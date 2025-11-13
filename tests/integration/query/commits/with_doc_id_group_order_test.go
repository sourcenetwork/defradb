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

func TestQueryCommitsOrderedAndGroupedByDocID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"Fred",
						"age":	25
					}`,
			},
			testUtils.Request{
				Request: ` {
					_commits(groupBy: [docID], order: {docID: DESC}) {
						docID
					}
				}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"docID": "bae-2487fd12-227f-582b-a7ed-3dd5d4b61fce",
						},
						{
							"docID": "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

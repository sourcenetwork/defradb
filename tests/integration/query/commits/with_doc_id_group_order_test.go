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
		Description: "Simple all commits query, grouped and ordered by docID",
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
					commits(groupBy: [docID], order: {docID: DESC}) {
						docID
					}
				}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"docID": "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7",
						},
						{
							"docID": "bae-2bb3e007-c40c-5264-8656-45e024cc4776",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

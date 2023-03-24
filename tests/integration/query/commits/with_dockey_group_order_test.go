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

func TestQueryCommitsOrderedAndGroupedByDocKey(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, grouped and ordered by dockey",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"John",
						"Age":	21
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"Fred",
						"Age":	25
					}`,
			},
			testUtils.Request{
				Request: ` {
					commits(groupBy: [dockey], order: {dockey: DESC}) {
						dockey
					}
				}`,
				Results: []map[string]any{
					{
						"dockey": "bae-b2103437-f5bd-52b6-99b1-5970412c5201",
					},
					{
						"dockey": "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

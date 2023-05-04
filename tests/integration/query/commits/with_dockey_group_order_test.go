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
					commits(groupBy: [dockey], order: {dockey: DESC}) {
						dockey
					}
				}`,
				Results: []map[string]any{
					{
						"dockey": "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
					},
					{
						"dockey": "bae-72f3dc53-1846-55d5-915c-28c4e83cc891",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

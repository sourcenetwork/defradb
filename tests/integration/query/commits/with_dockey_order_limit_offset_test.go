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

func TestQueryCommitsWithDockeyAndOrderAndLimitAndOffset(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order, limit and offset",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	23
				}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	24
				}`,
			},
			testUtils.Request{
				Request: `query {
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: ASC}, limit: 2, offset: 4) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeic2z67t72ty7op6aoqzpz7larpubb473naqipho7rftoivkmubh7a",
						"height": int64(2),
					},
					{
						"cid":    "bafybeids3sq532txo55elkc5eba7ns65s5lfbkx5x7cg2sxtpm253cyw5i",
						"height": int64(3),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

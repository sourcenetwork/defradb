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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsWithDocIDAndOrderAndLimitAndOffset(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.CreateDoc{
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
			&action.Request{
				Request: `query {
						_commits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738", order: {height: ASC}, limit: 2, offset: 4) {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":    "bafyreia4x5ju33jenbimdqbtnuqc7pby4lydpa7efyk5iu4nl6urm6ofla",
							"height": int64(2),
						},
						{
							"cid":    "bafyreiayx64xmsfgk2dz6mga2hcgm5ajbwrx2nhiroxyzdk7tfojjrl3fe",
							"height": int64(3),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

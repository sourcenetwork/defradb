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

func TestQueryCommitsWithDocIDAndOrderAndLimitAndOffset(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, order, limit and offset",
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
						commits(docID: "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7", order: {height: ASC}, limit: 2, offset: 4) {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid":    "bafyreido4fwolghako5ogh4jcy6tr3butjicfwubk27uyuimlm366rtdmy",
							"height": int64(2),
						},
						{
							"cid":    "bafyreiaa6r6d6ure3murz63ebfbmwlmtgm5rc5wbqvucufnku2k3vlgsga",
							"height": int64(3),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

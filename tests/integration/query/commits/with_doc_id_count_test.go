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

func TestQueryCommitsWithDocIDAndLinkCount(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple latest commits query with docID and link count",
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
						commits(docID: "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7") {
							cid
							_count(field: links)
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid":    "bafyreiae763hq5srsefplqrehpsuyieuwmbvblgzdma7srss522yciumhu",
							"_count": 0,
						},
						{
							"cid":    "bafyreiht7yhnnrgbwgyu5toe3exvpkovzrefzr6midu5secnlr546oel3q",
							"_count": 0,
						},
						{
							"cid":    "bafyreidtdklweht7ainl5rrdeqscr3cwr72sr4lehzrpmmnnbvnvstavnm",
							"_count": 2,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

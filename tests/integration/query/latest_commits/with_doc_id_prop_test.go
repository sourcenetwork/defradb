// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package latest_commits

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryLastCommitsWithDocIDProperty(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple latest commits query with docID property",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
			},
			testUtils.Request{
				Request: `query {
						latestCommits(docID: "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7") {
							docID
						}
					}`,
				Results: map[string]any{
					"latestCommits": []map[string]any{
						{
							"docID": "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

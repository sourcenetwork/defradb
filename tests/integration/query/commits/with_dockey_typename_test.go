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

func TestQueryCommitsWithDockeyWithTypeName(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with dockey and typename",
		Request: `query {
					commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
						cid
						__typename
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
			},
		},
		Results: []map[string]any{
			{
				"cid":        "bafybeidst2mzxhdoh4ayjdjoh4vibo7vwnuoxk3xgyk5mzmep55jklni2a",
				"__typename": "Commit",
			},
			{
				"cid":        "bafybeihhypcsqt7blkrqtcmpl43eo3yunrog5pchox5naji6hisdme4swm",
				"__typename": "Commit",
			},
			{
				"cid":        "bafybeih75uved5tzh5p7q4kzmkal7ezpjph7fnij6vn5krkgyc3jap2k5m",
				"__typename": "Commit",
			},
		},
	}

	executeTestCase(t, test)
}

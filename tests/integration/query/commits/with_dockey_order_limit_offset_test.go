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
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with dockey, order, limit and offset",
		Request: `query {
					commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {height: ASC}, limit: 2, offset: 4) {
						cid
						height
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
		Updates: map[int]map[int][]string{
			0: {
				0: {
					`{
						"Age": 22
					}`,
					`{
						"Age": 23
					}`,
					`{
						"Age": 24
					}`,
				},
			},
		},
		Results: []map[string]any{
			{
				"cid":    "bafybeidxeexqpsbf2qqrrkrysdztf2q5mqfwabwrcxdkjuolf6fsyzzyh4",
				"height": int64(2),
			},
			{
				"cid":    "bafybeifaxl4u5wmokgr4jviru6dz7teg7f2fomusxrvh7o5nh2a32jk3va",
				"height": int64(3),
			},
		},
	}

	executeTestCase(t, test)
}

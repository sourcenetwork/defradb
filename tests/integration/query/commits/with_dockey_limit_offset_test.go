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

func TestQueryCommitsWithDockeyAndLimitAndOffset(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with dockey, limit and offset",
		Query: `query {
					commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", limit: 2, offset: 1) {
						cid
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
				"cid": "bafybeiaxjhz6dna7fyf7tqo5hooilwvaezswd5xfsmb2lfgcy7tpzklikm",
			},
			{
				"cid": "bafybeicvef4ugls2dl7j4hibt2ahxss2i2i4bbgps7tkjiaoybp6q73mca",
			},
		},
	}

	executeTestCase(t, test)
}

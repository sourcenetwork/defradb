// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithGroupByWithGroupWithDocKey(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with DocKey filter on _group",
		Request: `query {
					Users(groupBy: [Age]) {
						Age
						_group(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
							Name
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				// bae-52b9170d-b77a-5887-b877-cbdbb99b009f
				`{
					"Name": "John",
					"Age": 21
				}`,
				`{
					"Name": "Bob",
					"Age": 32
				}`,
				`{
					"Name": "Fred",
					"Age": 21
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Age":    uint64(32),
				"_group": []map[string]any{},
			},
			{
				"Age": uint64(21),
				"_group": []map[string]any{
					{
						"Name": "John",
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

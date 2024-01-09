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

func TestQuerySimpleWithGroupByWithGroupWithDocIDs(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with docIDs filter on _group",
		Request: `query {
					Users(groupBy: [Age]) {
						Age
						_group(docIDs: ["bae-52b9170d-b77a-5887-b877-cbdbb99b009f", "bae-9b2e1434-9d61-5eb1-b3b9-82e8e40729a7"]) {
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
				// bae-9b2e1434-9d61-5eb1-b3b9-82e8e40729a7
				`{
					"Name": "Fred",
					"Age": 21
				}`,
				`{
					"Name": "Shahzad",
					"Age": 21
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Age":    int64(32),
				"_group": []map[string]any{},
			},
			{
				"Age": int64(21),
				"_group": []map[string]any{
					{
						"Name": "John",
					},
					{
						"Name": "Fred",
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

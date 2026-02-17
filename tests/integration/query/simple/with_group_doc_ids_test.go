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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithGroupByWithGroupWithDocIDs(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Fred",
					"Age": 21
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Shahzad",
					"Age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						GROUP(docID: ["bae-619ea0d2-35ba-5e8c-ac4d-2b769937213b", "bae-1b3c71c0-3632-58b6-9a6a-b3c72713e9fe"]) {
							Name
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(21),
							"GROUP": []map[string]any{
								{
									"Name": "John",
								},
								{
									"Name": "Fred",
								},
							},
						},
						{
							"Age":   int64(32),
							"GROUP": []map[string]any{},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

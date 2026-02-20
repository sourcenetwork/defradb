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

func TestQuerySimpleWithGroupByStringWithInnerGroupBooleanAndSumOfCount(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 25,
					"Verified": true
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32,
					"Verified": true
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 34,
					"Verified": false
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55,
					"Verified": true
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19,
					"Verified": false
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						SUM(GROUP: {field: COUNT})
						GROUP (groupBy: [Verified]){
							Verified
							COUNT(GROUP: {})
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"SUM":  int64(3),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"COUNT":    int(2),
								},
								{
									"Verified": false,
									"COUNT":    int(1),
								},
							},
						},
						{
							"Name": "Alice",
							"SUM":  int64(1),
							"GROUP": []map[string]any{
								{
									"Verified": false,
									"COUNT":    int(1),
								},
							},
						},
						{
							"Name": "Carlo",
							"SUM":  int64(1),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"COUNT":    int(1),
								},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

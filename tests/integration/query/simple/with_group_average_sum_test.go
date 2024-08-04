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

func TestQuerySimpleWithGroupByStringWithInnerGroupBooleanAndSumOfCountOfInt(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by string, with child group by boolean, and sum of average on int",
		Request: `query {
					Users(groupBy: [Name]) {
						Name
						_sum(_group: {field: _avg})
						_group (groupBy: [Verified]){
							Verified
							_avg(_group: {field: Age})
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 25,
					"Verified": true
				}`,
				`{
					"Name": "John",
					"Age": 32,
					"Verified": true
				}`,
				`{
					"Name": "John",
					"Age": 34,
					"Verified": false
				}`,
				`{
					"Name": "Carlo",
					"Age": 55,
					"Verified": true
				}`,
				`{
					"Name": "Alice",
					"Age": 19,
					"Verified": false
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"Name": "John",
					"_sum": float64(62.5),
					"_group": []map[string]any{
						{
							"Verified": true,
							"_avg":     float64(28.5),
						},
						{
							"Verified": false,
							"_avg":     float64(34),
						},
					},
				},
				{
					"Name": "Carlo",
					"_sum": float64(55),
					"_group": []map[string]any{
						{
							"Verified": true,
							"_avg":     float64(55),
						},
					},
				},
				{
					"Name": "Alice",
					"_sum": float64(19),
					"_group": []map[string]any{
						{
							"Verified": false,
							"_avg":     float64(19),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

// Note: this test should follow a different code path to `_avg` on it's own
// utilising the existing `_sum` node instead of adding a new one.  This test cannot
// verify that that code path is taken, but it does verfiy that the correct result
// is returned to the consumer in case the more efficient code path is taken.
func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildIntegerAverageAndSum(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by string, average and sum on non-rendered group integer value",
		Request: `query {
					Users(groupBy: [Name]) {
						Name
						_avg(_group: {field: Age})
						_sum(_group: {field: Age})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 32
				}`,
				`{
					"Name": "John",
					"Age": 38
				}`,
				// It is important to test negative values here, due to the auto-typing of numbers
				`{
					"Name": "Alice",
					"Age": -19
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"Name": "John",
					"_avg": float64(35),
					"_sum": int64(70),
				},
				{
					"Name": "Alice",
					"_avg": float64(-19),
					"_sum": int64(-19),
				},
			},
		},
	}

	executeTestCase(t, test)
}

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

// Note: this test should follow a different code path to `_avg` on it's own
// utilising the existing `_count` node instead of adding a new one.  This test cannot
// verify that that code path is taken, but it does verfiy that the correct result
// is returned to the consumer in case the more efficient code path is taken.
func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildIntegerAverageAndCount(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by string, average and sum on non-rendered group integer value",
		Request: `query {
					Users(groupBy: [Name]) {
						Name
						_avg(_group: {field: Age})
						_count(_group: {})
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
		Results: []map[string]any{
			{
				"Name":   "John",
				"_avg":   float64(35),
				"_count": int(2),
			},
			{
				"Name":   "Alice",
				"_avg":   float64(-19),
				"_count": int(1),
			},
		},
	}

	executeTestCase(t, test)
}

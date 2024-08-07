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

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildIntegerAverageWithLimitAndOffset(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by string, limited average on non-rendered group integer value",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 38
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 28
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": -19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						_avg(_group: {field: Age, offset: 1, limit: 2})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_avg": float64(35),
						},
						{
							"Name": "Alice",
							"_avg": float64(0),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

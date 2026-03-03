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

// Note: this test should follow a different code path to `AVG` on it's own
// utilising the existing `COUNT` node instead of adding a new one.  This test cannot
// verify that code path is taken, but it does verfiy that the correct result
// is returned to the consumer in case the more efficient code path is taken.
func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildIntegerAverageAndCount(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 38
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": -19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						AVG(GROUP: {field: Age})
						COUNT(GROUP: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":  "John",
							"AVG":   float64(35),
							"COUNT": int(2),
						},
						{
							"Name":  "Alice",
							"AVG":   float64(-19),
							"COUNT": int(1),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

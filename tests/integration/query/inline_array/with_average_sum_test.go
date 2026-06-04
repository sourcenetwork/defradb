// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package inline_array

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// Note: this test should follow a different code path to `AVG` on it's own
// utilising the existing `SUM` node instead of adding a new one.  This test cannot
// verify that code path is taken, but it does verfiy that the correct result
// is returned to the consumer in case the more efficient code path is taken.
func TestQueryInlineIntegerArrayWithAverageAndSum(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": [-1, 0, 9, 0]
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [name]) {
						name
						AVG(favouriteIntegers: {})
						SUM(favouriteIntegers: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"AVG":  float64(2),
							"SUM":  int64(8),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

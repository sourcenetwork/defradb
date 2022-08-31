// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package inline_array

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryInlineIntegerArrayWithSumWithLimit(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array, limited sum of integer array",
		Query: `query {
					users {
						Name
						_sum(FavouriteIntegers: {limit: 2})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "Shahzad",
					"FavouriteIntegers": [-1, 2, -1, 1, 0]
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Name": "Shahzad",
				"_sum": int64(1),
			},
		},
	}

	executeTestCase(t, test)
}

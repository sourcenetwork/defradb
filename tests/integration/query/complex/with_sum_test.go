// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package complex

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryComplexWithSumOnInlineAndManyToMany(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array with no filter, sum of integer array",
		Query: `query {
					publisher {
						name
						ThisMakesNoSenseToSumButHey: _sum(favouritePageNumbers: {})
						TotalRating: _sum(published: {field: rating})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "The Coffee Table Book",
					"rating": 4.9,
					"publisher_id": "bae-09468fb6-b7c6-57df-898e-8de473d114b3"
				}`,
			},
			2: {
				`{
					"name": "Pendant Publishing",
					"address": "600 Madison Ave., New York, New York",
					"favouritePageNumbers": [-1, 2, -1, 1, 0]
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"name":                        "Pendant Publishing",
				"ThisMakesNoSenseToSumButHey": int64(1),
				"TotalRating":                 float64(4.9),
			},
		},
	}

	executeTestCase(t, test)
}

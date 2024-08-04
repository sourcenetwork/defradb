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

func TestQuerySimpleWithAverageWithFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query, average with filter",
		Request: `query {
					_avg(Users: {field: Age, filter: {Age: {_gt: 26}}})
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
				`{
					"Name": "Bob",
					"Age": 30
				}`,
				`{
					"Name": "Alice",
					"Age": 32
				}`,
			},
		},
		Results: map[string]any{
			"_avg": float64(31),
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithAverageWithDateTimeFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query, average with datetime filter",
		Request: `query {
					_avg(Users: {field: Age, filter: {CreatedAt: {_gt: "2017-07-23T03:46:56-05:00"}}})
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21,
					"CreatedAt": "2017-07-23T03:46:56-05:00"
				}`,
				`{
					"Name": "Bob",
					"Age": 30,
					"CreatedAt": "2018-07-23T03:46:56-05:00"
				}`,
				`{
					"Name": "Alice",
					"Age": 32,
					"CreatedAt": "2019-07-23T03:46:56-05:00"
				}`,
			},
		},
		Results: map[string]any{
			"_avg": float64(31),
		},
	}

	executeTestCase(t, test)
}

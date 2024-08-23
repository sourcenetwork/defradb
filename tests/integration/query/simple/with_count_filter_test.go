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

func TestQuerySimpleWithCountWithFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query, count with filter",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 30
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 32
				}`,
			},
			testUtils.Request{
				Request: `query {
					_count(Users: {filter: {Age: {_gt: 26}}})
				}`,
				Results: map[string]any{
					"_count": 2,
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithCountWithDateTimeFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query, count with datetime filter",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21,
					"CreatedAt": "2017-07-23T03:46:56-05:00"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 30,
					"CreatedAt": "2017-09-23T03:46:56-05:00"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 32,
					"CreatedAt": "2017-10-23T03:46:56-05:00"
				}`,
			},
			testUtils.Request{
				Request: `query {
					_count(Users: {filter: {CreatedAt: {_gt: "2017-08-23T03:46:56-05:00"}}})
				}`,
				Results: map[string]any{
					"_count": 2,
				},
			},
		},
	}

	executeTestCase(t, test)
}

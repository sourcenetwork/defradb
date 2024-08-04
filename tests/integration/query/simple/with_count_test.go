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

func TestQuerySimpleWithCountOnUndefined(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query, count on undefined",
		Request: `query {
					_count
				}`,
		ExpectedError: "aggregate must be provided with a property to aggregate",
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithCountOnEmptyCollection(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query, count on empty",
		Request: `query {
					_count(Users: {})
				}`,
		Results: map[string]any{
			"_count": 0,
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithCount(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query, count",
		Request: `query {
					_count(Users: {})
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
			},
		},
		Results: map[string]any{
			"_count": 2,
		},
	}

	executeTestCase(t, test)
}

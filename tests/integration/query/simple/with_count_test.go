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
	test := testUtils.TestCase{
		Description: "Simple query, count on undefined",
		Actions: []any{
			testUtils.Request{
				Request: `query {
					_count
				}`,
				ExpectedError: "aggregate must be provided with a property to aggregate",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithCountOnEmptyCollection(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query, count on empty",
		Actions: []any{
			testUtils.Request{
				Request: `query {
					_count(Users: {})
				}`,
				Results: map[string]any{
					"_count": 0,
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithCount(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query, count",
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
			testUtils.Request{
				Request: `query {
					_count(Users: {})
				}`,
				Results: map[string]any{
					"_count": 2,
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithAliasedCount_OnEmptyCollection_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query, aliased count on empty",
		Actions: []any{
			testUtils.Request{
				Request: `query {
					number: _count(Users: {})
				}`,
				Results: map[string]any{
					"number": 0,
				},
			},
		},
	}

	executeTestCase(t, test)
}

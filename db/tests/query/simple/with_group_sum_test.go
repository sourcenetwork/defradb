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

// TODO!!!!! once scalar are merged, this should be capable of summing int/float arrays - likely needs some tweaks in generator.go and query.go

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/db/tests"
)

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndSumOfUndefined(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with sum on unspecified field",
		Query: `query {
					users (groupBy: [Name]) {
						Name
						_sum
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 32
			}`)},
		},
		ExpectedError: "Aggregate must be provided with a property to aggregate.",
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildIntegerSum(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with group by string, sum on non-rendered group integer value",
		Query: `query {
					users(groupBy: [Name]) {
						Name
						_sum(field: {_group: Age})
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 32
			}`),
				(`{
				"Name": "John",
				"Age": 38
			}`),
				// It is important to test negative values here, due to the auto-typing of numbers
				(`{
				"Name": "Alice",
				"Age": -19
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"Name": "John",
				"_sum": int64(70),
			},
			{
				"Name": "Alice",
				"_sum": int64(-19),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildNilSum(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with group by string, sum on non-rendered group nil and integer values",
		Query: `query {
					users(groupBy: [Name]) {
						Name
						_sum(field: {_group: Age})
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 32
			}`),
				// Age is undefined here
				(`{
				"Name": "John"
			}`),
				(`{
				"Name": "Alice",
				"Age": 19
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"Name": "Alice",
				"_sum": int64(19),
			},
			{
				"Name": "John",
				"_sum": int64(32),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildEmptyFloatSum(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with group by string, sum on non-rendered group float (default) value",
		Query: `query {
					users(groupBy: [Name]) {
						Name
						_sum(field: {_group: HeightM})
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"HeightM": 1.82
			}`),
				(`{
				"Name": "John",
				"HeightM": 1.89
			}`),
				(`{
				"Name": "Alice"
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"Name": "John",
				"_sum": float64(3.71),
			},
			{
				"Name": "Alice",
				"_sum": float64(0),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildFloatSum(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with group by string, sum on non-rendered group float value",
		Query: `query {
					users(groupBy: [Name]) {
						Name
						_sum(field: {_group: HeightM})
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"HeightM": 1.82
			}`),
				(`{
				"Name": "John",
				"HeightM": 1.89
			}`),
				(`{
				"Name": "Alice",
				"HeightM": 2.04
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"Name": "John",
				"_sum": float64(3.71),
			},
			{
				"Name": "Alice",
				"_sum": float64(2.04),
			},
		},
	}

	executeTestCase(t, test)
}

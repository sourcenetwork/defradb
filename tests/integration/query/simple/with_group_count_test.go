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

func TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndChildCount(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with group by number, no children, count on non-rendered group",
		Query: `query {
					users(groupBy: [Age]) {
						Age
						_count(_group: {})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 32
				}`,
				`{
					"Name": "Bob",
					"Age": 32
				}`,
				`{
					"Name": "Alice",
					"Age": 19
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Age":    uint64(32),
				"_count": 2,
			},
			{
				"Age":    uint64(19),
				"_count": 1,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndChildCountOnEmptyCollection(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with group by number, no children, count on non-rendered group, empty collection",
		Query: `query {
					users(groupBy: [Age]) {
						Age
						_count(_group: {})
					}
				}`,
		Results: []map[string]interface{}{},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithRenderedGroupAndChildCount(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with group by number, no children, count on rendered group",
		Query: `query {
					users(groupBy: [Age]) {
						Age
						_count(_group: {})
						_group {
							Name
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 32
				}`,
				`{
					"Name": "Bob",
					"Age": 32
				}`,
				`{
					"Name": "Alice",
					"Age": 19
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Age":    uint64(32),
				"_count": 2,
				"_group": []map[string]interface{}{
					{
						"Name": "Bob",
					},
					{
						"Name": "John",
					},
				},
			},
			{
				"Age":    uint64(19),
				"_count": 1,
				"_group": []map[string]interface{}{
					{
						"Name": "Alice",
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithUndefinedField(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with group by number, count on undefined field",
		Query: `query {
					users(groupBy: [Age]) {
						Age
						_count
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 32
				}`,
				`{
					"Name": "Bob",
					"Age": 32
				}`,
				`{
					"Name": "Alice",
					"Age": 19
				}`,
			},
		},
		ExpectedError: "Aggregate must be provided with a property to aggregate.",
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndAliasesChildCount(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with group by number, no children, aliased count on non-rendered group",
		Query: `query {
					users(groupBy: [Age]) {
						Age
						Count: _count(_group: {})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 32
				}`,
				`{
					"Name": "Bob",
					"Age": 32
				}`,
				`{
					"Name": "Alice",
					"Age": 19
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Age":   uint64(32),
				"Count": 2,
			},
			{
				"Age":   uint64(19),
				"Count": 1,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndDuplicatedAliasedChildCounts(
	t *testing.T,
) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with group by number, no children, duplicated aliased count on non-rendered group",
		Query: `query {
					users(groupBy: [Age]) {
						Age
						Count1: _count(_group: {})
						Count2: _count(_group: {})
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 32
				}`,
				`{
					"Name": "Bob",
					"Age": 32
				}`,
				`{
					"Name": "Alice",
					"Age": 19
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Age":    uint64(32),
				"Count1": 2,
				"Count2": 2,
			},
			{
				"Age":    uint64(19),
				"Count1": 1,
				"Count2": 1,
			},
		},
	}

	executeTestCase(t, test)
}

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

func TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndChildCountWithFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by number, no children, count on non-rendered, unfiltered group",
		Request: `query {
					Users(groupBy: [Age]) {
						Age
						_count(_group: {filter: {Age: {_gt: 26}}})
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
		Results: []map[string]any{
			{
				"Age":    int64(32),
				"_count": 2,
			},
			{
				"Age":    int64(19),
				"_count": 0,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithRenderedGroupAndChildCountWithFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by number, no children, count on non-rendered, filtered group",
		Request: `query {
					Users(groupBy: [Age]) {
						Age
						_count(_group: {filter: {Age: {_gt: 26}}})
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
		Results: []map[string]any{
			{
				"Age":    int64(32),
				"_count": 2,
				"_group": []map[string]any{
					{
						"Name": "Bob",
					},
					{
						"Name": "John",
					},
				},
			},
			{
				"Age":    int64(19),
				"_count": 0,
				"_group": []map[string]any{
					{
						"Name": "Alice",
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithRenderedGroupWithFilterAndChildCountWithMatchingFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by number, no children, count on non-rendered, matching filtered group",
		Request: `query {
					Users(groupBy: [Age]) {
						Age
						_count(_group: {filter: {Name: {_eq: "John"}}})
						_group(filter: {Name: {_eq: "John"}}) {
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
		Results: []map[string]any{
			{
				"Age":    int64(32),
				"_count": 1,
				"_group": []map[string]any{
					{
						"Name": "John",
					},
				},
			},
			{
				"Age":    int64(19),
				"_count": 0,
				"_group": []map[string]any{},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithRenderedGroupWithFilterAndChildCountWithDifferentFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by number, no children, count on non-rendered, different group filter",
		Request: `query {
					Users(groupBy: [Age]) {
						Age
						_count(_group: {filter: {Age: {_gt: 26}}})
						_group(filter: {Name: {_eq: "John"}}) {
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
		Results: []map[string]any{
			{
				"Age":    int64(32),
				"_count": 2,
				"_group": []map[string]any{
					{
						"Name": "John",
					},
				},
			},
			{
				"Age":    int64(19),
				"_count": 0,
				"_group": []map[string]any{},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndChildCountsWithDifferentFilters(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by number, no children, multiple counts on non-rendered, unfiltered group",
		Request: `query {
					Users(groupBy: [Age]) {
						Age
						C1: _count(_group: {filter: {Age: {_gt: 26}}})
						C2: _count(_group: {filter: {Age: {_lt: 26}}})
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
		Results: []map[string]any{
			{
				"Age": int64(32),
				"C1":  2,
				"C2":  0,
			},
			{
				"Age": int64(19),
				"C1":  0,
				"C2":  1,
			},
		},
	}

	executeTestCase(t, test)
}

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

func TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndChildSumWithFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by number, no children, sum on non-rendered, unfiltered group",
		Request: `query {
					Users(groupBy: [Age]) {
						Age
						_sum(_group: {field: Age, filter: {Age: {_gt: 26}}})
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
				"Age":  uint64(32),
				"_sum": int64(64),
			},
			{
				"Age":  uint64(19),
				"_sum": int64(0),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithRenderedGroupAndChildSumWithFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by number, no children, sum on rendered, unfiltered group",
		Request: `query {
					Users(groupBy: [Age]) {
						Age
						_sum(_group: {field: Age, filter: {Age: {_gt: 26}}})
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
				"Age":  uint64(32),
				"_sum": int64(64),
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
				"Age":  uint64(19),
				"_sum": int64(0),
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

func TestQuerySimpleWithGroupByNumberWithRenderedGroupWithFilterAndChildSumWithMatchingFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by number, no children, sum on rendered, matching filtered group",
		Request: `query {
					Users(groupBy: [Age]) {
						Age
						_sum(_group: {field: Age, filter: {Name: {_eq: "John"}}})
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
				"Age":  uint64(32),
				"_sum": int64(32),
				"_group": []map[string]any{
					{
						"Name": "John",
					},
				},
			},
			{
				"Age":    uint64(19),
				"_sum":   int64(0),
				"_group": []map[string]any{},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithRenderedGroupWithFilterAndChildSumWithDifferentFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by number, no children, sum on non-rendered, different filtered group",
		Request: `query {
					Users(groupBy: [Age]) {
						Age
						_sum(_group: {field: Age, filter: {Age: {_gt: 26}}})
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
				"Age":  uint64(32),
				"_sum": int64(64),
				"_group": []map[string]any{
					{
						"Name": "John",
					},
				},
			},
			{
				"Age":    uint64(19),
				"_sum":   int64(0),
				"_group": []map[string]any{},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndChildSumsWithDifferentFilters(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by number, no children, sum on non-rendered, unfiltered group",
		Request: `query {
					Users(groupBy: [Age]) {
						Age
						S1: _sum(_group: {field: Age, filter: {Age: {_gt: 26}}})
						S2: _sum(_group: {field: Age, filter: {Age: {_lt: 26}}})
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
				"Age": uint64(32),
				"S1":  int64(64),
				"S2":  int64(0),
			},
			{
				"Age": uint64(19),
				"S1":  int64(0),
				"S2":  int64(19),
			},
		},
	}

	executeTestCase(t, test)
}

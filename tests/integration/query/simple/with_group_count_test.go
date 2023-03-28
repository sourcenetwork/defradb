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

func TestQuerySimpleWithoutGroupByWithCountOnGroup(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query without group by, no children, count on non-existant group",
		Request: `query {
					Users {
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
			},
		},
		ExpectedError: "_group may only be referenced when within a groupBy request",
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithCountOnInnerNonExistantGroup(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query without group by, no children, count on inner non-existant group",
		Request: `query {
					Users(groupBy: [Age]) {
						Age
						_group {
							Name
							_count(_group: {})
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 32
				}`,
			},
		},
		ExpectedError: "_group may only be referenced when within a groupBy request",
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndChildCount(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by number, no children, count on non-rendered group",
		Request: `query {
					Users(groupBy: [Age]) {
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
		Results: []map[string]any{
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
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by number, no children, count on non-rendered group, empty collection",
		Request: `query {
					Users(groupBy: [Age]) {
						Age
						_count(_group: {})
					}
				}`,
		Results: []map[string]any{},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithRenderedGroupAndChildCount(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by number, no children, count on rendered group",
		Request: `query {
					Users(groupBy: [Age]) {
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
		Results: []map[string]any{
			{
				"Age":    uint64(32),
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
				"Age":    uint64(19),
				"_count": 1,
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

func TestQuerySimpleWithGroupByNumberWithUndefinedField(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by number, count on undefined field",
		Request: `query {
					Users(groupBy: [Age]) {
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
		ExpectedError: "aggregate must be provided with a property to aggregate",
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndAliasesChildCount(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by number, no children, aliased count on non-rendered group",
		Request: `query {
					Users(groupBy: [Age]) {
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
		Results: []map[string]any{
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
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by number, no children, duplicated aliased count on non-rendered group",
		Request: `query {
					Users(groupBy: [Age]) {
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
		Results: []map[string]any{
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

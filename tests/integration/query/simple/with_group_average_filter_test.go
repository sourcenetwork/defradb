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

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildAverageWithFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with group by string, no children, average on non-rendered, unfiltered group",
		Query: `query {
					users(groupBy: [Name]) {
						Name
						_avg(_group: {field: Age, filter: {Age: {_gt: 26}}})
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 34
			}`),
				(`{
				"Name": "John",
				"Age": 32
			}`),
				(`{
				"Name": "Alice",
				"Age": 19
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"Name": "Alice",
				"_avg": float64(0),
			},
			{
				"Name": "John",
				"_avg": float64(33),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithRenderedGroupAndChildAverageWithFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with group by string, no children, average on rendered, unfiltered group",
		Query: `query {
					users(groupBy: [Name]) {
						Name
						_avg(_group: {field: Age, filter: {Age: {_gt: 26}}})
						_group {
							Age
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 34
			}`),
				(`{
				"Name": "John",
				"Age": 32
			}`),
				(`{
				"Name": "Alice",
				"Age": 19
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"Name": "Alice",
				"_avg": float64(0),
				"_group": []map[string]interface{}{
					{
						"Age": uint64(19),
					},
				},
			},
			{
				"Name": "John",
				"_avg": float64(33),
				"_group": []map[string]interface{}{
					{
						"Age": uint64(32),
					},
					{
						"Age": uint64(34),
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithRenderedGroupWithFilterAndChildAverageWithMatchingFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with group by string, no children, average on rendered, matching filtered group",
		Query: `query {
					users(groupBy: [Name]) {
						Name
						_avg(_group: {field: Age, filter: {Age: {_gt: 33}}})
						_group(filter: {Age: {_gt: 33}}) {
							Age
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 34
			}`),
				(`{
				"Name": "John",
				"Age": 32
			}`),
				(`{
				"Name": "Alice",
				"Age": 19
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"Name":   "Alice",
				"_avg":   float64(0),
				"_group": []map[string]interface{}{},
			},
			{
				"Name": "John",
				"_avg": float64(34),
				"_group": []map[string]interface{}{
					{
						"Age": uint64(34),
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithRenderedGroupWithFilterAndChildAverageWithDifferentFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with group by string, no children, average on non-rendered, different filtered group",
		Query: `query {
					users(groupBy: [Name]) {
						Name
						_avg(_group: {field: Age, filter: {Age: {_gt: 33}}})
						_group(filter: {Age: {_lt: 33}}) {
							Age
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 34
			}`),
				(`{
				"Name": "John",
				"Age": 32
			}`),
				(`{
				"Name": "Alice",
				"Age": 19
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"Name": "Alice",
				"_avg": float64(0),
				"_group": []map[string]interface{}{
					{
						"Age": uint64(19),
					},
				},
			},
			{
				"Name": "John",
				"_avg": float64(34),
				"_group": []map[string]interface{}{
					{
						"Age": uint64(32),
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildAveragesWithDifferentFilters(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with group by string, no children, average on non-rendered, unfiltered group",
		Query: `query {
					users(groupBy: [Name]) {
						Name
						A1: _avg(_group: {field: Age, filter: {Age: {_gt: 26}}})
						A2: _avg(_group: {field: Age, filter: {Age: {_lt: 26}}})
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 34
			}`),
				(`{
				"Name": "John",
				"Age": 32
			}`),
				(`{
				"Name": "Alice",
				"Age": 19
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"Name": "Alice",
				"A1":   float64(0),
				"A2":   float64(19),
			},
			{
				"Name": "John",
				"A1":   float64(33),
				"A2":   float64(0),
			},
		},
	}

	executeTestCase(t, test)
}

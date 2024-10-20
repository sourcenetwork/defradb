// Copyright 2024 Democratized Data Foundation
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

func TestQuerySimple_WithGroupByNumberWithoutRenderedGroupAndChildMinWithFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by number, no children, min on non-rendered, unfiltered group",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						_min(_group: {field: Age, filter: {Age: {_gt: 26}}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age":  int64(32),
							"_min": int64(32),
						},
						{
							"Age":  int64(19),
							"_min": nil,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithGroupByNumberWithRenderedGroupAndChildMinWithFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by number, no children, min on rendered, unfiltered group",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						_min(_group: {field: Age, filter: {Age: {_gt: 26}}})
						_group {
							Name
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age":  int64(32),
							"_min": int64(32),
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
							"Age":  int64(19),
							"_min": nil,
							"_group": []map[string]any{
								{
									"Name": "Alice",
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithGroupByNumberWithRenderedGroupWithFilterAndChildMinWithMatchingFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by number, no children, min on rendered, matching filtered group",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						_min(_group: {field: Age, filter: {Name: {_eq: "John"}}})
						_group(filter: {Name: {_eq: "John"}}) {
							Name
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age":  int64(32),
							"_min": int64(32),
							"_group": []map[string]any{
								{
									"Name": "John",
								},
							},
						},
						{
							"Age":    int64(19),
							"_min":   nil,
							"_group": []map[string]any{},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithGroupByNumberWithRenderedGroupWithFilterAndChildMinWithDifferentFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by number, no children, min on non-rendered, different filtered group",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						_min(_group: {field: Age, filter: {Age: {_gt: 26}}})
						_group(filter: {Name: {_eq: "John"}}) {
							Name
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age":  int64(32),
							"_min": int64(32),
							"_group": []map[string]any{
								{
									"Name": "John",
								},
							},
						},
						{
							"Age":    int64(19),
							"_min":   nil,
							"_group": []map[string]any{},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithGroupByNumberWithoutRenderedGroupAndChildMinWithDifferentFilters_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by number, no children, min on non-rendered, unfiltered group",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						S1: _min(_group: {field: Age, filter: {Age: {_gt: 26}}})
						S2: _min(_group: {field: Age, filter: {Age: {_lt: 26}}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(32),
							"S1":  int64(32),
							"S2":  nil,
						},
						{
							"Age": int64(19),
							"S1":  nil,
							"S2":  int64(19),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

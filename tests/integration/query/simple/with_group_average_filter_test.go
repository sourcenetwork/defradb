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
	test := testUtils.TestCase{
		Description: "Simple query with group by string, no children, average on non-rendered, unfiltered group",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 34
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
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
					Users(groupBy: [Name]) {
						Name
						_avg(_group: {field: Age, filter: {Age: {_gt: 26}}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_avg": float64(33),
						},
						{
							"Name": "Alice",
							"_avg": float64(0),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithRenderedGroupAndChildAverageWithFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by string, no children, average on rendered, unfiltered group",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 34
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
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
					Users(groupBy: [Name]) {
						Name
						_avg(_group: {field: Age, filter: {Age: {_gt: 26}}})
						_group {
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_avg": float64(33),
							"_group": []map[string]any{
								{
									"Age": int64(34),
								},
								{
									"Age": int64(32),
								},
							},
						},
						{
							"Name": "Alice",
							"_avg": float64(0),
							"_group": []map[string]any{
								{
									"Age": int64(19),
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

func TestQuerySimpleWithGroupByStringWithRenderedGroupAndChildAverageWithDateTimeFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by string, no children, average on rendered, unfiltered group",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 34,
					"CreatedAt": "2019-07-23T03:46:56-05:00"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32,
					"CreatedAt": "2018-07-23T03:46:56-05:00"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19,
					"CreatedAt": "2011-07-23T03:46:56-05:00"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						_avg(_group: {field: Age, filter: {CreatedAt: {_gt: "2017-07-23T03:46:56-05:00"}}})
						_group {
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_avg": float64(33),
							"_group": []map[string]any{
								{
									"Age": int64(32),
								},
								{
									"Age": int64(34),
								},
							},
						},
						{
							"Name": "Alice",
							"_avg": float64(0),
							"_group": []map[string]any{
								{
									"Age": int64(19),
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

func TestQuerySimpleWithGroupByStringWithRenderedGroupWithFilterAndChildAverageWithMatchingFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by string, no children, average on rendered, matching filtered group",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 34
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
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
					Users(groupBy: [Name]) {
						Name
						_avg(_group: {field: Age, filter: {Age: {_gt: 33}}})
						_group(filter: {Age: {_gt: 33}}) {
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_avg": float64(34),
							"_group": []map[string]any{
								{
									"Age": int64(34),
								},
							},
						},
						{
							"Name":   "Alice",
							"_avg":   float64(0),
							"_group": []map[string]any{},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithRenderedGroupWithFilterAndChildAverageWithMatchingDateTimeFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by string, no children, average on rendered, matching datetime filtered group",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 34,
					"CreatedAt": "2017-07-23T03:46:56-05:00"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32,
					"CreatedAt": "2011-07-23T03:46:56-05:00"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19,
					"CreatedAt": "2010-07-23T03:46:56-05:00"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						_avg(_group: {field: Age, filter: {CreatedAt: {_gt: "2016-07-23T03:46:56-05:00"}}})
						_group(filter: {CreatedAt: {_gt: "2016-07-23T03:46:56-05:00"}}) {
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_avg": float64(34),
							"_group": []map[string]any{
								{
									"Age": int64(34),
								},
							},
						},
						{
							"Name":   "Alice",
							"_avg":   float64(0),
							"_group": []map[string]any{},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithRenderedGroupWithFilterAndChildAverageWithDifferentFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by string, no children, average on non-rendered, different filtered group",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 34
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
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
					Users(groupBy: [Name]) {
						Name
						_avg(_group: {field: Age, filter: {Age: {_gt: 33}}})
						_group(filter: {Age: {_lt: 33}}) {
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_avg": float64(34),
							"_group": []map[string]any{
								{
									"Age": int64(32),
								},
							},
						},
						{
							"Name": "Alice",
							"_avg": float64(0),
							"_group": []map[string]any{
								{
									"Age": int64(19),
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

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildAveragesWithDifferentFilters(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by string, no children, average on non-rendered, unfiltered group",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 34
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
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
					Users(groupBy: [Name]) {
						Name
						A1: _avg(_group: {field: Age, filter: {Age: {_gt: 26}}})
						A2: _avg(_group: {field: Age, filter: {Age: {_lt: 26}}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"A1":   float64(33),
							"A2":   float64(0),
						},
						{
							"Name": "Alice",
							"A1":   float64(0),
							"A2":   float64(19),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildAverageWithFilterAndNilItem(t *testing.T) {
	// This test checks that the appended/internal nil filter does not clash with the consumer-defined filter
	test := testUtils.TestCase{
		Description: "Simple query with group by string, no children, average with filter on non-rendered, unfiltered group",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 34
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 30
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John"
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
					Users(groupBy: [Name]) {
						Name
						_avg(_group: {field: Age, filter: {Age: {_lt: 33}}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_avg": float64(31),
						},
						{
							"Name": "Alice",
							"_avg": float64(19),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

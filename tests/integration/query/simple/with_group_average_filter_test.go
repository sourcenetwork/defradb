// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildAverageWithFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 34
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						AVG(GROUP: {field: Age, filter: {Age: {_gt: 26}}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"AVG":  float64(33),
						},
						{
							"Name": "Alice",
							"AVG":  float64(0),
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
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 34
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						AVG(GROUP: {field: Age, filter: {Age: {_gt: 26}}})
						GROUP {
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"AVG":  float64(33),
							"GROUP": []map[string]any{
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
							"AVG":  float64(0),
							"GROUP": []map[string]any{
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
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 34,
					"CreatedAt": "2019-07-23T03:46:56-05:00"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32,
					"CreatedAt": "2018-07-23T03:46:56-05:00"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19,
					"CreatedAt": "2011-07-23T03:46:56-05:00"
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						AVG(GROUP: {field: Age, filter: {CreatedAt: {_gt: "2017-07-23T03:46:56-05:00"}}})
						GROUP {
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"AVG":  float64(33),
							"GROUP": []map[string]any{
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
							"AVG":  float64(0),
							"GROUP": []map[string]any{
								{
									"Age": int64(19),
								},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithRenderedGroupWithFilterAndChildAverageWithMatchingFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 34
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						AVG(GROUP: {field: Age, filter: {Age: {_gt: 33}}})
						GROUP(filter: {Age: {_gt: 33}}) {
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"AVG":  float64(34),
							"GROUP": []map[string]any{
								{
									"Age": int64(34),
								},
							},
						},
						{
							"Name":  "Alice",
							"AVG":   float64(0),
							"GROUP": []map[string]any{},
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
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 34,
					"CreatedAt": "2017-07-23T03:46:56-05:00"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32,
					"CreatedAt": "2011-07-23T03:46:56-05:00"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19,
					"CreatedAt": "2010-07-23T03:46:56-05:00"
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						AVG(GROUP: {field: Age, filter: {CreatedAt: {_gt: "2016-07-23T03:46:56-05:00"}}})
						GROUP(filter: {CreatedAt: {_gt: "2016-07-23T03:46:56-05:00"}}) {
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"AVG":  float64(34),
							"GROUP": []map[string]any{
								{
									"Age": int64(34),
								},
							},
						},
						{
							"Name":  "Alice",
							"AVG":   float64(0),
							"GROUP": []map[string]any{},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithRenderedGroupWithFilterAndChildAverageWithDifferentFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 34
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						AVG(GROUP: {field: Age, filter: {Age: {_gt: 33}}})
						GROUP(filter: {Age: {_lt: 33}}) {
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"AVG":  float64(34),
							"GROUP": []map[string]any{
								{
									"Age": int64(32),
								},
							},
						},
						{
							"Name": "Alice",
							"AVG":  float64(0),
							"GROUP": []map[string]any{
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
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 34
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						A1: AVG(GROUP: {field: Age, filter: {Age: {_gt: 26}}})
						A2: AVG(GROUP: {field: Age, filter: {Age: {_lt: 26}}})
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
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 34
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 30
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						AVG(GROUP: {field: Age, filter: {Age: {_lt: 33}}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"AVG":  float64(31),
						},
						{
							"Name": "Alice",
							"AVG":  float64(19),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

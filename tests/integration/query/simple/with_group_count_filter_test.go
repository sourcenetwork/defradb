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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndChildCountWithFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						COUNT(GROUP: {filter: {Age: {_gt: 26}}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age":   int64(32),
							"COUNT": 2,
						},
						{
							"Age":   int64(19),
							"COUNT": 0,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithRenderedGroupAndChildCountWithFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						COUNT(GROUP: {filter: {Age: {_gt: 26}}})
						GROUP {
							Name
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age":   int64(32),
							"COUNT": 2,
							"GROUP": []map[string]any{
								{
									"Name": "Bob",
								},
								{
									"Name": "John",
								},
							},
						},
						{
							"Age":   int64(19),
							"COUNT": 0,
							"GROUP": []map[string]any{
								{
									"Name": "Alice",
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

func TestQuerySimpleWithGroupByNumberWithRenderedGroupWithFilterAndChildCountWithMatchingFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						COUNT(GROUP: {filter: {Name: {_eq: "John"}}})
						GROUP(filter: {Name: {_eq: "John"}}) {
							Name
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age":   int64(32),
							"COUNT": 1,
							"GROUP": []map[string]any{
								{
									"Name": "John",
								},
							},
						},
						{
							"Age":   int64(19),
							"COUNT": 0,
							"GROUP": []map[string]any{},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithRenderedGroupWithFilterAndChildCountWithDifferentFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						COUNT(GROUP: {filter: {Age: {_gt: 26}}})
						GROUP(filter: {Name: {_eq: "John"}}) {
							Name
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age":   int64(32),
							"COUNT": 2,
							"GROUP": []map[string]any{
								{
									"Name": "John",
								},
							},
						},
						{
							"Age":   int64(19),
							"COUNT": 0,
							"GROUP": []map[string]any{},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndChildCountsWithDifferentFilters(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						C1: COUNT(GROUP: {filter: {Age: {_gt: 26}}})
						C2: COUNT(GROUP: {filter: {Age: {_lt: 26}}})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
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
				},
			},
		},
	}

	executeTestCase(t, test)
}

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

func TestQuerySimpleWithGroupByStringWithGroupNumberFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by with child filter",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 25
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
					"Name": "Carlo",
					"Age": 55
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
						_group (filter: {Age: {_gt: 26}}){
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
							"_group": []map[string]any{
								{
									"Age": int64(55),
								},
							},
						},
						{
							"Name": "John",
							"_group": []map[string]any{
								{
									"Age": int64(32),
								},
							},
						},
						{
							"Name":   "Alice",
							"_group": []map[string]any{},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithGroupNumberWithParentFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by with number filter",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 25
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
					"Name": "Carlo",
					"Age": 55
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
					Users(groupBy: [Name], filter: {Age: {_gt: 26}}) {
						Name
						_group {
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
							"_group": []map[string]any{
								{
									"Age": int64(55),
								},
							},
						},
						{
							"Name": "John",
							"_group": []map[string]any{
								{
									"Age": int64(32),
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

func TestQuerySimpleWithGroupByStringWithUnrenderedGroupNumberWithParentFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by with number filter",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 25
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
					"Name": "Carlo",
					"Age": 55
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
					Users(groupBy: [Name], filter: {Age: {_gt: 26}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
						},
						{
							"Name": "John",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithInnerGroupBooleanThenInnerNumberFilterThatExcludesAll(
	t *testing.T,
) {
	test := testUtils.TestCase{
		Description: "Simple query with group by string, with child group by boolean, with child number filter that excludes all records",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 25,
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32,
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 34,
					"Verified": false
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55,
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19,
					"Verified": false
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						_group (groupBy: [Verified]){
							Verified
							_group (filter: {Age: {_gt: 260}}) {
								Age
							}
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_group": []map[string]any{
								{
									"Verified": true,
									"_group":   []map[string]any{},
								},
								{
									"Verified": false,
									"_group":   []map[string]any{},
								},
							},
						},
						{
							"Name": "Carlo",
							"_group": []map[string]any{
								{
									"Verified": true,
									"_group":   []map[string]any{},
								},
							},
						},
						{
							"Name": "Alice",
							"_group": []map[string]any{
								{
									"Verified": false,
									"_group":   []map[string]any{},
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

func TestQuerySimpleWithGroupByStringWithMultipleGroupNumberFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by with child filter",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 25
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
					"Name": "Carlo",
					"Age": 55
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
						G1: _group (filter: {Age: {_gt: 26}}){
							Age
						}
						G2: _group (filter: {Age: {_lt: 26}}){
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
							"G1": []map[string]any{
								{
									"Age": int64(55),
								},
							},
							"G2": []map[string]any{},
						},
						{
							"Name": "John",
							"G1": []map[string]any{
								{
									"Age": int64(32),
								},
							},
							"G2": []map[string]any{
								{
									"Age": int64(25),
								},
							},
						},
						{
							"Name": "Alice",
							"G1":   []map[string]any{},
							"G2": []map[string]any{
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

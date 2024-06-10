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

func TestQuerySimpleWithGroupByStringWithGroupNumberWithGroupOrder(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by string, and child order ascending",
		Request: `query {
					Users(groupBy: [Name]) {
						Name
						_group (order: {Age: ASC}){
							Age
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 25
				}`,
				`{
					"Name": "John",
					"Age": 32
				}`,
				`{
					"Name": "Carlo",
					"Age": 55
				}`,
				`{
					"Name": "Alice",
					"Age": 19
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "Carlo",
				"_group": []map[string]any{
					{
						"Age": int64(55),
					},
				},
			},
			{
				"Name": "Alice",
				"_group": []map[string]any{
					{
						"Age": int64(19),
					},
				},
			},
			{
				"Name": "John",
				"_group": []map[string]any{
					{
						"Age": int64(25),
					},
					{
						"Age": int64(32),
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithGroupNumberWithGroupOrderDescending(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by string, and child order descending",
		Request: `query {
					Users(groupBy: [Name]) {
						Name
						_group (order: {Age: DESC}){
							Age
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 25
				}`,
				`{
					"Name": "John",
					"Age": 32
				}`,
				`{
					"Name": "Carlo",
					"Age": 55
				}`,
				`{
					"Name": "Alice",
					"Age": 19
				}`,
			},
		},
		Results: []map[string]any{
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
					{
						"Age": int64(25),
					},
				},
			},
			{
				"Name": "Alice",
				"_group": []map[string]any{
					{
						"Age": int64(19),
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringAndOrderDescendingWithGroupNumberWithGroupOrder(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by string, and child order ascending",
		Request: `query {
					Users(groupBy: [Name], order: {Name: DESC}) {
						Name
						_group (order: {Age: ASC}){
							Age
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 25
				}`,
				`{
					"Name": "John",
					"Age": 32
				}`,
				`{
					"Name": "Carlo",
					"Age": 55
				}`,
				`{
					"Name": "Alice",
					"Age": 19
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "John",
				"_group": []map[string]any{
					{
						"Age": int64(25),
					},
					{
						"Age": int64(32),
					},
				},
			},
			{
				"Name": "Carlo",
				"_group": []map[string]any{
					{
						"Age": int64(55),
					},
				},
			},
			{
				"Name": "Alice",
				"_group": []map[string]any{
					{
						"Age": int64(19),
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithInnerGroupBooleanThenInnerOrderDescending(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by string, with child group by boolean, with child order desc",
		Request: `query {
					Users(groupBy: [Name]) {
						Name
						_group (groupBy: [Verified]){
							Verified
							_group (order: {Age: DESC}) {
								Age
							}
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 25,
					"Verified": true
				}`,
				`{
					"Name": "John",
					"Age": 32,
					"Verified": true
				}`,
				`{
					"Name": "John",
					"Age": 34,
					"Verified": false
				}`,
				`{
					"Name": "Carlo",
					"Age": 55,
					"Verified": true
				}`,
				`{
					"Name": "Alice",
					"Age": 19,
					"Verified": false
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "John",
				"_group": []map[string]any{
					{
						"Verified": true,
						"_group": []map[string]any{
							{
								"Age": int64(32),
							},
							{
								"Age": int64(25),
							},
						},
					},
					{
						"Verified": false,
						"_group": []map[string]any{
							{
								"Age": int64(34),
							},
						},
					},
				},
			},
			{
				"Name": "Carlo",
				"_group": []map[string]any{
					{
						"Verified": true,
						"_group": []map[string]any{
							{
								"Age": int64(55),
							},
						},
					},
				},
			},
			{
				"Name": "Alice",
				"_group": []map[string]any{
					{
						"Verified": false,
						"_group": []map[string]any{
							{
								"Age": int64(19),
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithInnerGroupBooleanAndOrderAscendingThenInnerOrderDescending(
	t *testing.T,
) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by string, with child group by boolean, with child order desc",
		Request: `query {
					Users(groupBy: [Name]) {
						Name
						_group (groupBy: [Verified], order: {Verified: ASC}){
							Verified
							_group (order: {Age: DESC}) {
								Age
							}
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 25,
					"Verified": false
				}`,
				`{
					"Name": "John",
					"Age": 32,
					"Verified": true
				}`,
				`{
					"Name": "John",
					"Age": 34,
					"Verified": false
				}`,
				`{
					"Name": "Carlo",
					"Age": 55,
					"Verified": true
				}`,
				`{
					"Name": "Alice",
					"Age": 19,
					"Verified": false
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "John",
				"_group": []map[string]any{
					{
						"Verified": false,
						"_group": []map[string]any{
							{
								"Age": int64(34),
							},
							{
								"Age": int64(25),
							},
						},
					},
					{
						"Verified": true,
						"_group": []map[string]any{
							{
								"Age": int64(32),
							},
						},
					},
				},
			},
			{
				"Name": "Alice",
				"_group": []map[string]any{
					{
						"Verified": false,
						"_group": []map[string]any{
							{
								"Age": int64(19),
							},
						},
					},
				},
			},
			{
				"Name": "Carlo",
				"_group": []map[string]any{
					{
						"Verified": true,
						"_group": []map[string]any{
							{
								"Age": int64(55),
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

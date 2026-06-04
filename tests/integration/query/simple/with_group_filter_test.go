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

func TestQuerySimpleWithGroupByStringWithGroupNumberFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 25
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
					"Name": "Carlo",
					"Age": 55
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
						GROUP (filter: {Age: {_gt: 26}}){
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"GROUP": []map[string]any{
								{
									"Age": int64(32),
								},
							},
						},
						{
							"Name": "Carlo",
							"GROUP": []map[string]any{
								{
									"Age": int64(55),
								},
							},
						},
						{
							"Name":  "Alice",
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

func TestQuerySimpleWithGroupByStringWithGroupNumberWithParentFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 25
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
					"Name": "Carlo",
					"Age": 55
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
					Users(groupBy: [Name], filter: {Age: {_gt: 26}}) {
						Name
						GROUP {
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
							"GROUP": []map[string]any{
								{
									"Age": int64(55),
								},
							},
						},
						{
							"Name": "John",
							"GROUP": []map[string]any{
								{
									"Age": int64(32),
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

func TestQuerySimpleWithGroupByStringWithUnrenderedGroupNumberWithParentFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 25
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
					"Name": "Carlo",
					"Age": 55
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
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithInnerGroupBooleanThenInnerNumberFilterThatExcludesAll(
	t *testing.T,
) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 25,
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32,
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 34,
					"Verified": false
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55,
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19,
					"Verified": false
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						GROUP (groupBy: [Verified]){
							Verified
							GROUP (filter: {Age: {_gt: 260}}) {
								Age
							}
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"GROUP":    []map[string]any{},
								},
								{
									"Verified": false,
									"GROUP":    []map[string]any{},
								},
							},
						},
						{
							"Name": "Alice",
							"GROUP": []map[string]any{
								{
									"Verified": false,
									"GROUP":    []map[string]any{},
								},
							},
						},
						{
							"Name": "Carlo",
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"GROUP":    []map[string]any{},
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

func TestQuerySimpleWithGroupByStringWithMultipleGroupNumberFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 25
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
					"Name": "Carlo",
					"Age": 55
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
						G1: GROUP (filter: {Age: {_gt: 26}}){
							Age
						}
						G2: GROUP (filter: {Age: {_lt: 26}}){
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
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
							"Name": "Carlo",
							"G1": []map[string]any{
								{
									"Age": int64(55),
								},
							},
							"G2": []map[string]any{},
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
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

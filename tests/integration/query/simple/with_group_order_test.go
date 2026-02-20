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

func TestQuerySimpleWithGroupByStringWithGroupNumberWithGroupOrder(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 25
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
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
					Users(groupBy: [Name]) {
						Name
						GROUP (order: {Age: ASC}){
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
									"Age": int64(25),
								},
								{
									"Age": int64(32),
								},
							},
						},
						{
							"Name": "Alice",
							"GROUP": []map[string]any{
								{
									"Age": int64(19),
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
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithGroupNumberWithGroupOrderDescending(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 25
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
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
					Users(groupBy: [Name]) {
						Name
						GROUP (order: {Age: DESC}){
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
								{
									"Age": int64(25),
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
							"Name": "Alice",
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

func TestQuerySimpleWithGroupByStringAndOrderDescendingWithGroupNumberWithGroupOrder(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 25
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
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
					Users(groupBy: [Name], order: {Name: DESC}) {
						Name
						GROUP (order: {Age: ASC}){
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
									"Age": int64(25),
								},
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
							"Name": "Alice",
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

func TestQuerySimpleWithGroupByStringWithInnerGroupBooleanThenInnerOrderDescending(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 25,
					"Verified": true
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32,
					"Verified": true
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 34,
					"Verified": false
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55,
					"Verified": true
				}`,
			},
			&action.CreateDoc{
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
							GROUP (order: {Age: DESC}) {
								Age
							}
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"GROUP": []map[string]any{
										{
											"Age": int64(55),
										},
									},
								},
							},
						},
						{
							"Name": "John",
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"GROUP": []map[string]any{
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
									"GROUP": []map[string]any{
										{
											"Age": int64(34),
										},
									},
								},
							},
						},
						{
							"Name": "Alice",
							"GROUP": []map[string]any{
								{
									"Verified": false,
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
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithInnerGroupBooleanAndOrderAscendingThenInnerOrderDescending(
	t *testing.T,
) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 25,
					"Verified": false
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32,
					"Verified": true
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 34,
					"Verified": false
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55,
					"Verified": true
				}`,
			},
			&action.CreateDoc{
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
						GROUP (groupBy: [Verified], order: {Verified: ASC}){
							Verified
							GROUP (order: {Age: DESC}) {
								Age
							}
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"GROUP": []map[string]any{
										{
											"Age": int64(55),
										},
									},
								},
							},
						},
						{
							"Name": "John",
							"GROUP": []map[string]any{
								{
									"Verified": false,
									"GROUP": []map[string]any{
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
									"GROUP": []map[string]any{
										{
											"Age": int64(32),
										},
									},
								},
							},
						},
						{
							"Name": "Alice",
							"GROUP": []map[string]any{
								{
									"Verified": false,
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
			},
		},
	}

	executeTestCase(t, test)
}

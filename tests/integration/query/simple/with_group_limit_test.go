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

func TestQuerySimpleWithGroupByNumberWithGroupLimit(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
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
					Users(groupBy: [Age]) {
						Age
						GROUP(limit: 1) {
							Name
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(32),
							"GROUP": []map[string]any{
								{
									"Name": "John",
								},
							},
						},
						{
							"Age": int64(19),
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

func TestQuerySimpleWithGroupByNumberWithMultipleGroupsWithDifferentLimits(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
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
					Users(groupBy: [Age]) {
						Age
						G1: GROUP(limit: 1) {
							Name
						}
						G2: GROUP(limit: 2) {
							Name
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(32),
							"G1": []map[string]any{
								{
									"Name": "John",
								},
							},
							"G2": []map[string]any{
								{
									"Name": "Bob",
								},
								{
									"Name": "John",
								},
							},
						},
						{
							"Age": int64(19),
							"G1": []map[string]any{
								{
									"Name": "Alice",
								},
							},
							"G2": []map[string]any{
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

func TestQuerySimpleWithGroupByNumberWithLimitAndGroupWithHigherLimit(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
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
					Users(groupBy: [Age], limit: 1) {
						Age
						GROUP(limit: 2) {
							Name
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(32),
							"GROUP": []map[string]any{
								{
									"Name": "Bob",
								},
								{
									"Name": "John",
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

func TestQuerySimpleWithGroupByNumberWithLimitAndGroupWithLowerLimit(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 42
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Age], limit: 2) {
						Age
						GROUP(limit: 1) {
							Name
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(42),
							"GROUP": []map[string]any{
								{
									"Name": "Alice",
								},
							},
						},
						{
							"Age": int64(32),
							"GROUP": []map[string]any{
								{
									"Name": "John",
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

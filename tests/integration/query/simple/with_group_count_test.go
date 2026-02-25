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

func TestQuerySimpleWithoutGroupByWithCountOnGroup(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						Age
						COUNT(GROUP: {})
					}
				}`,
				ExpectedError: "group may only be referenced when within a groupBy request",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithCountOnInnerNonExistantGroup(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						GROUP {
							Name
							COUNT(GROUP: {})
						}
					}
				}`,
				ExpectedError: "group may only be referenced when within a groupBy request",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndChildCount(t *testing.T) {
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
						COUNT(GROUP: {})
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
							"COUNT": 1,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndChildCountOnEmptyCollection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						COUNT(GROUP: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithRenderedGroupAndChildCount(t *testing.T) {
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
						COUNT(GROUP: {})
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
							"COUNT": 1,
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

func TestQuerySimpleWithGroupByNumberWithUndefinedField(t *testing.T) {
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
						COUNT
					}
				}`,
				ExpectedError: "aggregate must be provided with a property to aggregate",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndAliasesChildCount(t *testing.T) {
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
						Count: COUNT(GROUP: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age":   int64(32),
							"Count": 2,
						},
						{
							"Age":   int64(19),
							"Count": 1,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndDuplicatedAliasedChildCounts(
	t *testing.T,
) {
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
						Count1: COUNT(GROUP: {})
						Count2: COUNT(GROUP: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age":    int64(32),
							"Count1": 2,
							"Count2": 2,
						},
						{
							"Age":    int64(19),
							"Count1": 1,
							"Count2": 1,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

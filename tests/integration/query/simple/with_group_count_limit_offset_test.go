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

func TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndChildCountWithLimitAndOffset(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by number, no children, count with limit and offset on non-rendered group",
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
						_count(_group: {offset: 1, limit: 1})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age":    int64(32),
							"_count": 1,
						},
						{
							"Age":    int64(19),
							"_count": 0,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithRenderedGroupWithLimitAndChildCountWithLimitAndOffset(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by number, child limit, count with limit and offset on rendered group",
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
					"Name": "Shahzad",
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
						_count(_group: {offset: 1, limit: 1})
						_group (limit: 2) {
							Name
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age":    int64(32),
							"_count": 1,
							"_group": []map[string]any{
								{
									"Name": "Bob",
								},
								{
									"Name": "Shahzad",
								},
							},
						},
						{
							"Age":    int64(19),
							"_count": 0,
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

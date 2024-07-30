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

func TestQuerySimpleWithGroupByNumberWithGroupLimitAndOffset(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by number, no children, rendered, limited and offset group",
		Request: `query {
					Users(groupBy: [Age]) {
						Age
						_group(limit: 1, offset: 1) {
							Name
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 32
				}`,
				`{
					"Name": "Bob",
					"Age": 32
				}`,
				`{
					"Name": "Alice",
					"Age": 19
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"Age": int64(32),
					"_group": []map[string]any{
						{
							"Name": "John",
						},
					},
				},
				{
					"Age":    int64(19),
					"_group": []map[string]any{},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByNumberWithLimitAndOffsetAndWithGroupLimitAndOffset(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with group by number with limit and offset, no children, rendered, limited and offset group",
		Request: `query {
					Users(groupBy: [Age], limit: 1, offset: 1) {
						Age
						_group(limit: 1, offset: 1) {
							Name
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 32
				}`,
				`{
					"Name": "Bob",
					"Age": 32
				}`,
				`{
					"Name": "Alice",
					"Age": 19
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"Age":    int64(19),
					"_group": []map[string]any{},
				},
			},
		},
	}

	executeTestCase(t, test)
}

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

func TestQuerySimpleWithGroupByWithTypeName(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query group by and parent typename",
		Request: `query {
					Users(groupBy: [Name]) {
						Name
						__typename
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John"
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name":       "John",
				"__typename": "Users",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByWithChildTypeName(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query group by and child typename",
		Request: `query {
					Users(groupBy: [Name]) {
						Name
						_group {
							__typename
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John"
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "John",
				"_group": []map[string]any{
					{
						"__typename": "Users",
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

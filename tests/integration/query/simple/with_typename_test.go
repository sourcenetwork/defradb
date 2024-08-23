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

func TestQuerySimpleWithTypeName(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with typename",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						Name
						__typename
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":       "John",
							"__typename": "Users",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithAliasedTypeName(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with aliased typename",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						Name
						__typename
						t1: __typename
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":       "John",
							"__typename": "Users",
							"t1":         "Users",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

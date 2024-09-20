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

func TestQuerySimpleWithIntEqualToXOrYFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with logical compound filter (or)",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
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
					Users(filter: {_or: [{Age: {_eq: 55}}, {Age: {_eq: 19}}]}) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
							"Age":  int64(55),
						},
						{
							"Name": "Alice",
							"Age":  int64(19),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithInlineIntArray_EqualToXOrYFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with logical compound filter (or) on inline int array",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `type Users {
					Name: String
					FavoriteNumbers: [Int!]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"FavoriteNumbers": [10, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"FavoriteNumbers": [30, 40]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {_or: [
						{FavoriteNumbers: {_any: {_le: 100}}},
						{FavoriteNumbers: {_any: {_ge: 0}}},
					]}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Alice",
						},
						{
							"Name": "Bob",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

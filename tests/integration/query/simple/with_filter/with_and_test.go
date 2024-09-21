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

func TestQuerySimpleWithIntGreaterThanAndIntLessThanFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with logical compound filter (and)",
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
					Users(filter: {_and: [{Age: {_gt: 20}}, {Age: {_lt: 50}}]}) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"Age":  int64(21),
						},
						{
							"Name": "Bob",
							"Age":  int64(32),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithInlineIntArray_GreaterThanAndLessThanFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with logical compound filter (and) on inline int array",
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
					"FavoriteNumbers": [0, 10, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"FavoriteNumbers": [30, 40, 50]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {_and: [
						{FavoriteNumbers: {_all: {_ge: 0}}},
						{FavoriteNumbers: {_all: {_lt: 30}}},
					]}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
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

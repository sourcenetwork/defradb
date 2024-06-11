// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_one

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToOneWithCountWithCompoundOrFilterThatIncludesRelation(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation with count with _or filter that includes relation",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: bookAuthorGQLSchema,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Some Book",
					"rating": 4.0
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Some Other Book",
					"rating": 3.5
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Yet Another Book",
					"rating": 3.0
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "John Grisham",
					"age":          65,
					"verified":     true,
					"published_id": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Some Writer",
					"age":          45,
					"verified":     false,
					"published_id": testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Some Other Writer",
					"age":          35,
					"verified":     false,
					"published_id": testUtils.NewDocIndex(0, 2),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Yet Another Writer",
					"age":          30,
					"verified":     false,
					"published_id": testUtils.NewDocIndex(0, 3),
				},
			},
			testUtils.Request{
				Request: `query {
					_count(Book: {filter: {_or: [
						{_not: {author: {age: {_lt: 65}}} },
						{_not: {author: {age: {_gt: 30}}} }
					]}})
				}`,
				Results: []map[string]any{
					{
						"_count": int(2),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

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

// This test documents a bug and should be altered with:
// https://github.com/sourcenetwork/defradb/issues/1869
func TestQueryOneToOneWithCountWithCompoundOrFilterThatIncludesRelation(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation with count with _or filter that includes relation",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: bookAuthorGQLSchema,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				Doc: `{
					"name": "Painted House",
					"rating": 4.9
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-437092f3-7817-555c-bf8a-cc1c5a0a0db6
				Doc: `{
					"name": "Some Book",
					"rating": 4.0
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-1c890922-ddf9-5820-a888-c7f977848934
				Doc: `{
					"name": "Some Other Book",
					"rating": 3.5
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// TestQueryOneToOneWithCompoundOrFilterThatIncludesRelation
				Doc: `{
					"name": "Yet Another Book",
					"rating": 3.0
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Some Writer",
					"age": 45,
					"verified": false,
					"published_id": "bae-437092f3-7817-555c-bf8a-cc1c5a0a0db6"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Some Other Writer",
					"age": 35,
					"verified": false,
					"published_id": "bae-1c890922-ddf9-5820-a888-c7f977848934"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Yet Another Writer",
					"age": 30,
					"verified": false,
					"published_id": "TestQueryOneToOneWithCompoundOrFilterThatIncludesRelation"
				}`,
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
						"_count": "2",
					},
				},
			},
		},
	}

	testUtils.AssertPanic(
		t,
		func() {
			testUtils.ExecuteTestCase(t, test)
		},
	)
}

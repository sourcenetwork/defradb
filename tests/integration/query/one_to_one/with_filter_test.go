// Copyright 2022 Democratized Data Foundation
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

func TestQueryOneToOneWithNumericFilterOnParent(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation query with simple filter on sub type",
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
				CollectionID: 1,
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book {
						name
						rating
						author(filter: {age: {_eq: 65}}) {
							name
							age
						}
					}
				}`,
				Results: []map[string]any{
					{
						"name":   "Painted House",
						"rating": 4.9,
						"author": map[string]any{
							"name": "John Grisham",
							"age":  int64(65),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneWithStringFilterOnChild(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation query with simple filter on parent",
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
				CollectionID: 1,
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book(filter: {name: {_eq: "Painted House"}}) {
						name
						rating
						author {
							name
							age
						}
					}
				}`,
				Results: []map[string]any{
					{
						"name":   "Painted House",
						"rating": 4.9,
						"author": map[string]any{
							"name": "John Grisham",
							"age":  int64(65),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneWithBooleanFilterOnChild(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation query with simple sub filter on child",
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
				CollectionID: 1,
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book(filter: {author: {verified: {_eq: true}}}) {
						name
						rating
						author {
							name
							age
						}
					}
				}`,
				Results: []map[string]any{
					{
						"name":   "Painted House",
						"rating": 4.9,
						"author": map[string]any{
							"name": "John Grisham",
							"age":  int64(65),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneWithFilterThroughChildBackToParent(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation query with filter on parent referencing parent through child",
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
				// bae-d432bdfb-787d-5a1c-ac29-dc025ab80095
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8
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
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false,
					"published_id": "bae-d432bdfb-787d-5a1c-ac29-dc025ab80095"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book(filter: {author: {published: {rating: {_eq: 4.9}}}}) {
						name
						rating
						author {
							name
							age
						}
					}
				}`,
				Results: []map[string]any{
					{
						"name":   "Painted House",
						"rating": 4.9,
						"author": map[string]any{
							"name": "John Grisham",
							"age":  int64(65),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneWithBooleanFilterOnChildWithNoSubTypeSelection(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation with simple sub filter on child, but not child selections",
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
				CollectionID: 1,
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book(filter: {author: {verified: {_eq: true}}}) {
						name
						rating
					}
				}`,
				Results: []map[string]any{{
					"name":   "Painted House",
					"rating": 4.9,
				}},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneWithCompoundAndFilterThatIncludesRelation(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation with _and filter that includes relation",
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
				// bae-500a9445-bd90-580e-9191-d2d0ec1a5cf5
				Doc: `{
					"name": "Some Other Book",
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
					"age": 30,
					"verified": true,
					"published_id": "bae-500a9445-bd90-580e-9191-d2d0ec1a5cf5"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book(filter: {_and: [
						{rating: {_ge: 4.0}},
						{author: {verified: {_eq: true}}}
					]}) {
						name
						rating
					}
				}`,
				Results: []map[string]any{{
					"name":   "Painted House",
					"rating": 4.9,
				}},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneWithCompoundOrFilterThatIncludesRelation(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation with _or filter that includes relation",
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
					Book(filter: {_or: [
						{_and: [
							{rating: {_ge: 4.0}},
							{author: {age: {_le: 45}}}
						]},
						{_and: [
							{rating: {_le: 3.5}},
							{author: {age: {_ge: 35}}}
						]}
					]}) {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Some Other Book",
					},
					{
						"name": "Some Book",
					},
				},
			},
			testUtils.Request{
				Request: `query {
					Book(filter: {_or: [
						{_not: {author: {age: {_lt: 65}}} },
						{_not: {author: {age: {_gt: 30}}} }
					]}) {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Yet Another Book",
					},
					{
						"name": "Painted House",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_many

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToManyWithNumericGreaterThanFilterOnParent(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from the many side, simple filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: bookAuthorGQLSchema,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			testUtils.Request{
				Request: `query {
					Author(filter: {age: {_gt: 63}}) {
						name
						age
						published {
							name
							rating
						}
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John Grisham",
						"age":  uint64(65),
						"published": []map[string]any{
							{
								"name":   "Painted House",
								"rating": 4.9,
							},
							{
								"name":   "A Time for Mercy",
								"rating": 4.5,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithNumericGreaterThanChildFilterOnParentWithUnrenderedChild(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from the many side, simple filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: bookAuthorGQLSchema,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			testUtils.Request{
				Request: `query {
					Author(filter: {published: {rating: {_gt: 4.8}}, age: {_gt: 63}}) {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John Grisham",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithNumericGreaterThanFilterOnParentAndChild(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from the many side, simple filter on root and sub type",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: bookAuthorGQLSchema,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			testUtils.Request{
				Request: `query {
					Author(filter: {age: {_gt: 63}}) {
						name
						age
						published(filter: {rating: {_gt: 4.6}}) {
							name
							rating
						}
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John Grisham",
						"age":  uint64(65),
						"published": []map[string]any{
							{
								"name":   "Painted House",
								"rating": 4.9,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithMultipleAliasedFilteredChildren(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from the many side, simple filter on root and sub type",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: bookAuthorGQLSchema,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			testUtils.Request{
				Request: `query {
					Author {
						name
						age
						p1: published(filter: {rating: {_gt: 4.6}}) {
							name
							rating
						}
						p2: published(filter: {rating: {_lt: 4.6}}) {
							name
							rating
						}
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John Grisham",
						"age":  uint64(65),
						"p1": []map[string]any{
							{
								"name":   "Painted House",
								"rating": 4.9,
							},
						},
						"p2": []map[string]any{
							{
								"name":   "A Time for Mercy",
								"rating": 4.5,
							},
						},
					},
					{
						"name": "Cornelia Funke",
						"age":  uint64(62),
						"p1": []map[string]any{
							{
								"name":   "Theif Lord",
								"rating": 4.8,
							},
						},
						"p2": []map[string]any{},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithCompoundOperatorInFilterAndRelation(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query filter with compound operator and relation",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: bookAuthorGQLSchema,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			testUtils.Request{
				Request: `query {
					Author(filter: {_or: [
						{ name: {_eq: "Not existing author"}},
						{ _and: [
							{age: {_gt: 64}},
							{published: {rating: {_gt: 4.8}}}
						]},
						{ _and: [
							{age: {_gt: 80}},
							{published: {rating: {_lt: 4.9}}}
						]},
					]}) {
						name
					}
				}`,
				Results: []map[string]any{{
					"name": "John Grisham",
				}},
			},
			testUtils.Request{
				Request: `query {
					Author(filter: {_or: [
						{ name: {_eq: "Not existing author"}},
						{ _and: [
							{age: {_gt: 80}},
							{published: {rating: {_gt: 4.8}}}
						]},
						{ _and: [
							{age: {_lt: 64}},
							{published: {rating: {_lt: 4.9}}}
						]},
					]}) {
						name
					}
				}`,
				Results: []map[string]any{{
					"name": "Cornelia Funke",
				}},
			},
			testUtils.Request{
				Request: `query {
					Author(filter: {_or: [
						{ name: {_eq: "Not existing author"}},
						{ _and: [
							{age: {_gt: 30}},
							{published: {rating: {_eq: 4.5}}}
						]},
						{ _and: [
							{age: {_gt: 80}},
							{published: {rating: {_lt: 4.9}}}
						]},
					]}) {
						name
					}
				}`,
				Results: []map[string]any{{
					"name": "John Grisham",
				}},
			},
			testUtils.Request{
				Request: `query {
					Author(filter: {_or: [
						{ name: {_eq: "Not existing author"}},
						{ _and: [
							{age: {_gt: 30}},
							{published: {rating: {_eq: 4.8}}}
						]},
						{ _and: [
							{age: {_gt: 80}},
							{published: {rating: {_lt: 4.9}}}
						]},
					]}) {
						name
					}
				}`,
				Results: []map[string]any{{
					"name": "Cornelia Funke",
				}},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

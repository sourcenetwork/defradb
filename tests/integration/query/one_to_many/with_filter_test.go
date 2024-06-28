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
				// bae-be6d8024-4953-5a92-84b4-f042d25230c6
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-72e8c691-9f20-55e7-9228-8af1cf54cace"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-72e8c691-9f20-55e7-9228-8af1cf54cace
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
						"age":  int64(65),
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
				// bae-be6d8024-4953-5a92-84b4-f042d25230c6
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-72e8c691-9f20-55e7-9228-8af1cf54cace"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-72e8c691-9f20-55e7-9228-8af1cf54cace
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
				// bae-be6d8024-4953-5a92-84b4-f042d25230c6
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-72e8c691-9f20-55e7-9228-8af1cf54cace"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-72e8c691-9f20-55e7-9228-8af1cf54cace
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
						"age":  int64(65),
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
				// bae-be6d8024-4953-5a92-84b4-f042d25230c6
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-72e8c691-9f20-55e7-9228-8af1cf54cace"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-72e8c691-9f20-55e7-9228-8af1cf54cace
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
						"name": "Cornelia Funke",
						"age":  int64(62),
						"p1": []map[string]any{
							{
								"name":   "Theif Lord",
								"rating": 4.8,
							},
						},
						"p2": []map[string]any{},
					},
					{
						"name": "John Grisham",
						"age":  int64(65),
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
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-72e8c691-9f20-55e7-9228-8af1cf54cace"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Lord of the Rings",
					"rating": 5.0,
					"author_id": "bae-6bf29c1c-7112-5f4f-bfae-1c039479acf6"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-72e8c691-9f20-55e7-9228-8af1cf54cace
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-6bf29c1c-7112-5f4f-bfae-1c039479acf6
				Doc: `{
					"name": "John Tolkien",
					"age": 70,
					"verified": true
				}`,
			},
			testUtils.Request{
				Request: `query {
					Author(filter: {_or: [
						{_and: [
							{published: {rating: {_lt: 5.0}}},
							{published: {rating: {_gt: 4.8}}}
						]},
						{_and: [
							{age: {_le: 65}},
							{published: {name: {_like: "%Lord%"}}}
						]},
					]}) {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Cornelia Funke",
					},
					{
						"name": "John Grisham",
					},
				},
			},
			testUtils.Request{
				Request: `query {
					Author(filter: {_and: [
						{ _not: {published: {rating: {_gt: 4.8}}}},
						{ _not: {published: {rating: {_lt: 4.8}}}}
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

func TestQueryOneToMany_WithCompoundOperatorInFilterAndRelationAndCaseInsensitiveLike_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query filter with compound operator and relation",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: bookAuthorGQLSchema,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Tolkien",
					"age": 70,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Painted House",
					"rating":    4.9,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"rating":    4.5,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"rating":    4.8,
					"author_id": testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Lord of the Rings",
					"rating":    5.0,
					"author_id": testUtils.NewDocIndex(1, 2),
				},
			},
			testUtils.Request{
				Request: `query {
					Author(filter: {_or: [
						{_and: [
							{published: {rating: {_lt: 5.0}}},
							{published: {rating: {_gt: 4.8}}}
						]},
						{_and: [
							{age: {_le: 65}},
							{published: {name: {_ilike: "%lord%"}}}
						]},
					]}) {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Cornelia Funke",
					},
					{
						"name": "John Grisham",
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

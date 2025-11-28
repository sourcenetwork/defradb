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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToManyWithNumericGreaterThanFilterOnParent(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: bookAuthorGQLSchema,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-3d5a3204-4e55-5236-992a-ce27da27902b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-9d52c335-c8e3-5782-8daa-e359c106e0ab
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-3d5a3204-4e55-5236-992a-ce27da27902b
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
				Results: map[string]any{
					"Author": []map[string]any{
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
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithNumericGreaterThanChildFilterOnParentWithUnrenderedChild(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: bookAuthorGQLSchema,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-3d5a3204-4e55-5236-992a-ce27da27902b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-9d52c335-c8e3-5782-8daa-e359c106e0ab
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-3d5a3204-4e55-5236-992a-ce27da27902b
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
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John Grisham",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithNumericGreaterThanFilterOnParentAndChild(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: bookAuthorGQLSchema,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-3d5a3204-4e55-5236-992a-ce27da27902b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-9d52c335-c8e3-5782-8daa-e359c106e0ab
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-3d5a3204-4e55-5236-992a-ce27da27902b
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
				Results: map[string]any{
					"Author": []map[string]any{
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithMultipleAliasedFilteredChildren(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: bookAuthorGQLSchema,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-3d5a3204-4e55-5236-992a-ce27da27902b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-9d52c335-c8e3-5782-8daa-e359c106e0ab
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-3d5a3204-4e55-5236-992a-ce27da27902b
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
				Results: map[string]any{
					"Author": []map[string]any{
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
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithCompoundOperatorInFilterAndRelation(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: bookAuthorGQLSchema,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-3d5a3204-4e55-5236-992a-ce27da27902b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Lord of the Rings",
					"rating": 5.0,
					"author_id": "bae-3027a2d8-0820-5db3-a25f-20239a3571c8"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-9d52c335-c8e3-5782-8daa-e359c106e0ab
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-3d5a3204-4e55-5236-992a-ce27da27902b
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-3027a2d8-0820-5db3-a25f-20239a3571c8
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
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John Grisham",
						},
						{
							"name": "Cornelia Funke",
						},
					},
				},
				NonOrderedResults: true,
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
				Results: map[string]any{
					"Author": []map[string]any{
						{"name": "Cornelia Funke"},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToMany_WithCompoundOperatorInFilterAndRelationAndCaseInsensitiveLike_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
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
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John Grisham",
						},
						{
							"name": "Cornelia Funke",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToMany_WithAliasFilterOnRelated_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: bookAuthorGQLSchema,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-3d5a3204-4e55-5236-992a-ce27da27902b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-9d52c335-c8e3-5782-8daa-e359c106e0ab
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-3d5a3204-4e55-5236-992a-ce27da27902b
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			testUtils.Request{
				Request: `query {
					Author(filter: {_alias: {books: {rating: {_gt: 4.8}}}}) {
						name
						age
						books: published {
							name
							rating
						}
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John Grisham",
							"age":  int64(65),
							"books": []map[string]any{
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
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

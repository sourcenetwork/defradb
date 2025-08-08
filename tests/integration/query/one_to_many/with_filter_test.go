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
		Description: "One-to-many relation query from the many side, simple filter",
		Actions: []any{
			&action.AddSchema{
				Schema: bookAuthorGQLSchema,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-818aecea-02f9-5064-9e17-c8b7cc20e63f
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-f62bb529-3508-529d-8098-f97f9b67824c"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-f62bb529-3508-529d-8098-f97f9b67824c
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithNumericGreaterThanChildFilterOnParentWithUnrenderedChild(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from the many side, simple filter",
		Actions: []any{
			&action.AddSchema{
				Schema: bookAuthorGQLSchema,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-818aecea-02f9-5064-9e17-c8b7cc20e63f
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-f62bb529-3508-529d-8098-f97f9b67824c"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-f62bb529-3508-529d-8098-f97f9b67824c
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
		Description: "One-to-many relation query from the many side, simple filter on root and sub type",
		Actions: []any{
			&action.AddSchema{
				Schema: bookAuthorGQLSchema,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-818aecea-02f9-5064-9e17-c8b7cc20e63f
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-f62bb529-3508-529d-8098-f97f9b67824c"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-f62bb529-3508-529d-8098-f97f9b67824c
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
		Description: "One-to-many relation query from the many side, simple filter on root and sub type",
		Actions: []any{
			&action.AddSchema{
				Schema: bookAuthorGQLSchema,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-818aecea-02f9-5064-9e17-c8b7cc20e63f
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-f62bb529-3508-529d-8098-f97f9b67824c"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-f62bb529-3508-529d-8098-f97f9b67824c
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithCompoundOperatorInFilterAndRelation(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query filter with compound operator and relation",
		Actions: []any{
			&action.AddSchema{
				Schema: bookAuthorGQLSchema,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-f62bb529-3508-529d-8098-f97f9b67824c"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Lord of the Rings",
					"rating": 5.0,
					"author_id": "bae-eb11c625-3e66-56ac-8407-d543ca0c21f9"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-f62bb529-3508-529d-8098-f97f9b67824c
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-eb11c625-3e66-56ac-8407-d543ca0c21f9
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
		Description: "One-to-many relation query filter with compound operator and relation",
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
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToMany_WithAliasFilterOnRelated_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from the many side, alias filter",
		Actions: []any{
			&action.AddSchema{
				Schema: bookAuthorGQLSchema,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-818aecea-02f9-5064-9e17-c8b7cc20e63f
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-f62bb529-3508-529d-8098-f97f9b67824c"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-f62bb529-3508-529d-8098-f97f9b67824c
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package one_to_one

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToOneWithNumericFilterOnParent(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Filtering the related author subtype by an integer field returns the matching book.",
		Actions: []any{
			&action.AddCollection{
				SDL: bookAuthorGQLSchema,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "John Grisham",
					"age":          65,
					"verified":     true,
					"_publishedID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
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
				Results: map[string]any{
					"Book": []map[string]any{
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneWithStringFilterOnChild(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Filtering the root Book collection by a string field returns the matching book with its author.",
		Actions: []any{
			&action.AddCollection{
				SDL: bookAuthorGQLSchema,
			},
			&action.AddDoc{
				CollectionID: 0,
				// bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25
				Doc: `{
					"name": "Painted House",
					"rating": 4.9
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				// bae-9e70648f-c722-5875-97f5-574ec6f703e9
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"_publishedID": "bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25"
				}`,
			},
			&action.Request{
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
				Results: map[string]any{
					"Book": []map[string]any{
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneWithBooleanFilterOnChild(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Filtering books by a boolean field on the related author returns the matching result.",
		Actions: []any{
			&action.AddCollection{
				SDL: bookAuthorGQLSchema,
			},
			&action.AddDoc{
				CollectionID: 0,
				// bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25
				Doc: `{
					"name": "Painted House",
					"rating": 4.9
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				// bae-9e70648f-c722-5875-97f5-574ec6f703e9
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"_publishedID": "bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25"
				}`,
			},
			&action.Request{
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
				Results: map[string]any{
					"Book": []map[string]any{
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneWithFilterThroughChildBackToParent(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Filtering books by a field on the author's linked book traverses the relation back to parent.",
		Actions: []any{
			&action.AddCollection{
				SDL: bookAuthorGQLSchema,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "John Grisham",
					"age":          65,
					"verified":     true,
					"_publishedID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Cornelia Funke",
					"age":          62,
					"verified":     false,
					"_publishedID": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.Request{
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
				Results: map[string]any{
					"Book": []map[string]any{
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneWithBooleanFilterOnChildWithNoSubTypeSelection(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Filtering books by a boolean on the related author without selecting author fields returns books only.",
		Actions: []any{
			&action.AddCollection{
				SDL: bookAuthorGQLSchema,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "John Grisham",
					"age":          65,
					"verified":     true,
					"_publishedID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: `query {
					Book(filter: {author: {verified: {_eq: true}}}) {
						name
						rating
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":   "Painted House",
							"rating": 4.9,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneWithCompoundAndFilterThatIncludesRelation(t *testing.T) {
	test := testUtils.TestCase{
		Description: "A compound _and filter combining a book rating and a related author field narrows results correctly.",
		Actions: []any{
			&action.AddCollection{
				SDL: bookAuthorGQLSchema,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Some Book",
					"rating": 4.0
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Some Other Book",
					"rating": 3.0
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "John Grisham",
					"age":          65,
					"verified":     true,
					"_publishedID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Some Writer",
					"age":          45,
					"verified":     false,
					"_publishedID": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Some Other Writer",
					"age":          30,
					"verified":     true,
					"_publishedID": testUtils.NewDocIndex(0, 2),
				},
			},
			&action.Request{
				Request: `query {
					Book(filter: {_and: [
						{rating: {_geq: 4.0}},
						{author: {verified: {_eq: true}}}
					]}) {
						name
						rating
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":   "Painted House",
							"rating": 4.9,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneWithCompoundOrFilterThatIncludesRelation(t *testing.T) {
	test := testUtils.TestCase{
		Description: "A compound _or filter combining book rating and related author age returns the expected subset.",
		Actions: []any{
			&action.AddCollection{
				SDL: bookAuthorGQLSchema,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Some Book",
					"rating": 4.0
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Some Other Book",
					"rating": 3.5
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Yet Another Book",
					"rating": 3.0
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "John Grisham",
					"age":          65,
					"verified":     true,
					"_publishedID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Some Writer",
					"age":          45,
					"verified":     false,
					"_publishedID": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Some Other Writer",
					"age":          35,
					"verified":     false,
					"_publishedID": testUtils.NewDocIndex(0, 2),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Yet Another Writer",
					"age":          30,
					"verified":     false,
					"_publishedID": testUtils.NewDocIndex(0, 3),
				},
			},
			&action.Request{
				Request: `query {
					Book(filter: {_or: [
						{_and: [
							{rating: {_geq: 4.0}},
							{author: {age: {_leq: 45}}}
						]},
						{_and: [
							{rating: {_leq: 3.5}},
							{author: {age: {_geq: 35}}}
						]}
					]}) {
						name
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name": "Some Other Book",
						},
						{
							"name": "Some Book",
						},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request: `query {
					Book(filter: {_or: [
						{_not: {author: {age: {_lt: 65}}} },
						{_not: {author: {age: {_gt: 30}}} }
					]}) {
						name
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name": "Yet Another Book",
						},
						{
							"name": "Painted House",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOne_WithCompoundFiltersThatIncludesRelation_ShouldReturnResults(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Multiple compound filter shapes (_or, _and, _not) referencing a relation all return correct books.",
		Actions: []any{
			&action.AddCollection{
				SDL: bookAuthorGQLSchema,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Some Book",
					"rating": 4.0
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Some Other Book",
					"rating": 3.0
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "John Grisham",
					"age":          65,
					"verified":     true,
					"_publishedID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Some Writer",
					"age":          45,
					"verified":     false,
					"_publishedID": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Some Other Writer",
					"age":          30,
					"verified":     true,
					"_publishedID": testUtils.NewDocIndex(0, 2),
				},
			},
			&action.Request{
				Request: `query {
					Book(filter: {_or: [
						{rating: {_gt: 4.0}},
						{author: {age: {_eq: 30}}}
					]}) {
						name
						rating
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":   "Painted House",
							"rating": 4.9,
						},
						{
							"name":   "Some Other Book",
							"rating": 3.0,
						},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request: `query {
					Book(filter: {_and: [
						{rating: {_geq: 4.0}},
						{author: {age: {_eq: 45}}}
					]}) {
						name
						rating
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":   "Some Book",
							"rating": 4.0,
						},
					},
				},
			},
			&action.Request{
				// This is the same as {_not: {_and: [{rating: {_geq: 4.0}}, {author: {age: {_eq: 45}}}]}}
				Request: `query {
					Book(filter: {_not: {
						rating: {_geq: 4.0},
						author: {age: {_eq: 45}}
					}}) {
						name
						rating
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":   "Painted House",
							"rating": 4.9,
						},
						{
							"name":   "Some Other Book",
							"rating": 3.0,
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

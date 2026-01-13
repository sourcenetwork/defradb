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

func TestQueryOneToOneWithChildBooleanOrderDescending(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
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
					"name": "Theif Lord",
					"rating": 4.8
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"_publishedID": "bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false,
					"_publishedID": "bae-9793af00-a131-5ef2-b2c9-22b8053a11e7"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book(order: {author: {verified: DESC}}) {
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
						{
							"name":   "Theif Lord",
							"rating": 4.8,
							"author": map[string]any{
								"name": "Cornelia Funke",
								"age":  int64(62),
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToOneWithChildBooleanOrderAscending(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
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
					"name": "Theif Lord",
					"rating": 4.8
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"_publishedID": "bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false,
					"_publishedID": "bae-9793af00-a131-5ef2-b2c9-22b8053a11e7"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book(order: {author: {verified: ASC}}) {
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
							"name":   "Theif Lord",
							"rating": 4.8,
							"author": map[string]any{
								"name": "Cornelia Funke",
								"age":  int64(62),
							},
						},
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

	executeTestCase(t, test)
}

func TestQueryOneToOneWithChildIntOrderDescendingWithNoSubTypeFieldsSelected(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
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
					"name": "Theif Lord",
					"rating": 4.8
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"_publishedID": "bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false,
					"_publishedID": "bae-9793af00-a131-5ef2-b2c9-22b8053a11e7"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book(order: {author: {age: DESC}}) {
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
							"name":   "Theif Lord",
							"rating": 4.8,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToOneWithChildIntOrderAscendingWithNoSubTypeFieldsSelected(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
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
					"name": "Theif Lord",
					"rating": 4.8
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"_publishedID": "bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false,
					"_publishedID": "bae-9793af00-a131-5ef2-b2c9-22b8053a11e7"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book(order: {author: {age: ASC}}) {
						name
						rating
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":   "Theif Lord",
							"rating": 4.8,
						},
						{
							"name":   "Painted House",
							"rating": 4.9,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToOne_WithAliasedChildIntOrderAscending_ShouldOrder(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
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
					"name": "Theif Lord",
					"rating": 4.8
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "John Grisham",
					"age":          65,
					"verified":     true,
					"_publishedID": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Cornelia Funke",
					"age":          62,
					"verified":     false,
					"_publishedID": testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.Request{
				Request: `query {
					Book(order: {_alias: {writer: {age: ASC}}}) {
						name
						rating
						writer: author {
							age
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":   "Theif Lord",
							"rating": 4.8,
							"writer": map[string]any{
								"age": int64(62),
							},
						},
						{
							"name":   "Painted House",
							"rating": 4.9,
							"writer": map[string]any{
								"age": int64(65),
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToOne_WithChildAliasedIntOrderAscending_ShouldOrder(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
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
					"name": "Theif Lord",
					"rating": 4.8
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "John Grisham",
					"age":          65,
					"verified":     true,
					"_publishedID": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Cornelia Funke",
					"age":          62,
					"verified":     false,
					"_publishedID": testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.Request{
				Request: `query {
					Book(order: {author: {_alias: {authorAge: ASC}}}) {
						name
						rating
						author {
							authorAge: age
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":   "Theif Lord",
							"rating": 4.8,
							"author": map[string]any{
								"authorAge": int64(62),
							},
						},
						{
							"name":   "Painted House",
							"rating": 4.9,
							"author": map[string]any{
								"authorAge": int64(65),
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

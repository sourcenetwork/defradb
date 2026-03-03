// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package kind

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCollectionVersionUpdatesAddFieldKindForeignObject_WithAddCollectionCreatingOneToManyRelationToExistingCollection_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Author {
						name: String
					}
				`,
			},
			&action.AddCollection{
				SDL: `
					type Book {
						name: String
						author: Author
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John Grisham",
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":   "Painted House",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":   "A Time for Mercy",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: `query {
					Book {
						name
						_authorID
						author {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":      "A Time for Mercy",
							"_authorID": testUtils.NewDocIndex(0, 0),
							"author": map[string]any{
								"name": "John Grisham",
							},
						},
						{
							"name":      "Painted House",
							"_authorID": testUtils.NewDocIndex(0, 0),
							"author": map[string]any{
								"name": "John Grisham",
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

func TestCollectionVersionUpdatesAddFieldKindForeignObject_WithAddCollectionCreatingOneToManyRelationsToMultipleExistingCollections_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Author {
						name: String
					}
				`,
			},
			&action.AddCollection{
				SDL: `
					type Publisher {
						name: String
					}
				`,
			},
			&action.AddCollection{
				SDL: `
					type Book {
						name: String
						author: Author
						publisher: Publisher
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John Grisham",
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name": "Penguin Books",
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":      "Painted House",
					"author":    testUtils.NewDocIndex(0, 0),
					"publisher": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.Request{
				Request: `query {
					Book {
						name
						_authorID
						author {
							name
						}
						_publisherID
						publisher {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":      "Painted House",
							"_authorID": testUtils.NewDocIndex(0, 0),
							"author": map[string]any{
								"name": "John Grisham",
							},
							"_publisherID": testUtils.NewDocIndex(1, 0),
							"publisher": map[string]any{
								"name": "Penguin Books",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesAddFieldKindForeignObject_WithPatchAddingOneToManyRelationAfterSeparateAddCollections_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						name: String
					}
				`,
			},
			&action.AddCollection{
				SDL: `
					type Author {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Book/Fields/-", "value": {
							"Name": "author", "Kind": "Author", "RelationName": "author_book", "IsPrimary": true
						}},
						{ "op": "add", "path": "/Book/Fields/-", "value": {
							"Name": "_authorID", "Kind": 1, "RelationName": "author_book", "IsPrimary": true
						}},
						{ "op": "add", "path": "/Author/Fields/-", "value": {
							"Name": "books", "Kind": "[Book]", "RelationName": "author_book"
						}}
					]
				`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name": "John Grisham",
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":   "Painted House",
					"author": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":   "A Time to Kill",
					"author": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.Request{
				Request: `query {
					Author {
						name
						books {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John Grisham",
							"books": []map[string]any{
								{"name": "Painted House"},
								{"name": "A Time to Kill"},
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

func TestCollectionVersionUpdatesAddFieldKindForeignObject_WithMixedBatchHavingRelationToExistingAndNewCollections_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Author {
						name: String
					}
				`,
			},
			&action.AddCollection{
				SDL: `
					type Publisher {
						name: String
					}
					type Book {
						name: String
						author: Author
						publisher: Publisher
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John Grisham",
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name": "Penguin Books",
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":      "Painted House",
					"author":    testUtils.NewDocIndex(0, 0),
					"publisher": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.Request{
				Request: `query {
					Book {
						name
						author {
							name
						}
						publisher {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name": "Painted House",
							"author": map[string]any{
								"name": "John Grisham",
							},
							"publisher": map[string]any{
								"name": "Penguin Books",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesAddFieldKindForeignObject_WithChainedOneToManyRelationsAcrossSeparateCollections_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Publisher {
						name: String
					}
				`,
			},
			&action.AddCollection{
				SDL: `
					type Author {
						name: String
						publisher: Publisher
					}
				`,
			},
			&action.AddCollection{
				SDL: `
					type Book {
						name: String
						author: Author
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "Penguin Books",
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "John Grisham",
					"publisher": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":   "Painted House",
					"author": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.Request{
				Request: `query {
					Book {
						name
						author {
							name
							publisher {
								name
							}
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name": "Painted House",
							"author": map[string]any{
								"name": "John Grisham",
								"publisher": map[string]any{
									"name": "Penguin Books",
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

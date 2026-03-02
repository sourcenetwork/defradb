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

func TestCollectionVersionUpdatesAddFieldKindForeignObject(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": 16} }
					]
				`,
				ExpectedError: "no type found for given name. Type: 16",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesAddFieldKindForeignObject_UnknownCollection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo", "Kind": "Unknown"
						}}
					]
				`,
				ExpectedError: "no type found for given name. Field: foo, Kind: Unknown",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesAddFieldKindForeignObject_IDFieldMissingKind(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo", "Kind": "Users"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "_fooID"} }
					]
				`,
				ExpectedError: "relational id field of invalid kind. Field: _fooID, Expected: ID, Actual: 0",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesAddFieldKindForeignObject_IDFieldInvalidKind(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo", "Kind": "Users"
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "_fooID", "Kind": 2} }
					]
				`,
				ExpectedError: "relational id field of invalid kind. Field: _fooID, Expected: ID, Actual: Boolean",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesAddFieldKindForeignObject_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo", "Kind": "Users", "RelationName": "users_users", "IsPrimary": true
						}},
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "_fooID", "Kind": 1, "RelationName": "users_users", "IsPrimary": true
						}}
					]
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Keenan",
					"foo":  testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: `query {
					Users {
						name
						foo {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"foo":  nil,
						},
						{
							"name": "Keenan",
							"foo": map[string]any{
								"name": "John",
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

func TestCollectionVersionUpdatesAddFieldKindForeignObject_WithPatchAddingOneToOneRelationInSameBatch_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Author {
						name: String
					}
					type Book {
						title: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Author/Fields/-", "value": {
							"Name": "published", "Kind": "Book", "RelationName": "author_book", "IsPrimary": true
						}},
						{ "op": "add", "path": "/Author/Fields/-", "value": {
							"Name": "_publishedID", "Kind": 1, "RelationName": "author_book", "IsPrimary": true
						}},
						{ "op": "add", "path": "/Book/Fields/-", "value": {
							"Name": "author", "Kind": "Author", "RelationName": "author_book"
						}}
					]
				`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title": "The Great Gatsby",
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "F. Scott Fitzgerald",
					"published": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.Request{
				Request: `query {
					Author {
						name
						published {
							title
						}
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "F. Scott Fitzgerald",
							"published": map[string]any{
								"title": "The Great Gatsby",
							},
						},
					},
				},
			},
			&action.Request{
				Request: `query {
					Book {
						title
						author {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"title": "The Great Gatsby",
							"author": map[string]any{
								"name": "F. Scott Fitzgerald",
							},
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesAddFieldKindForeignObject_WithPatchAddingOneToManyRelationInSameBatch_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Author {
						name: String
					}
					type Book {
						title: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Author/Fields/-", "value": {
							"Name": "published", "Kind": "[Book]", "RelationName": "author_book"
						}},
						{ "op": "add", "path": "/Book/Fields/-", "value": {
							"Name": "author", "Kind": "Author", "RelationName": "author_book", "IsPrimary": true
						}},
						{ "op": "add", "path": "/Book/Fields/-", "value": {
							"Name": "_authorID", "Kind": 1, "RelationName": "author_book", "IsPrimary": true
						}}
					]
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "F. Scott Fitzgerald",
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "The Great Gatsby",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Tender Is the Night",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: `query {
					Author {
						name
						published {
							title
						}
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "F. Scott Fitzgerald",
							"published": []map[string]any{
								{"title": "The Great Gatsby"},
								{"title": "Tender Is the Night"},
							},
						},
					},
				},
			},
			&action.Request{
				Request: `query {
					Book {
						title
						author {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"title": "The Great Gatsby",
							"author": map[string]any{
								"name": "F. Scott Fitzgerald",
							},
						},
						{
							"title": "Tender Is the Night",
							"author": map[string]any{
								"name": "F. Scott Fitzgerald",
							},
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

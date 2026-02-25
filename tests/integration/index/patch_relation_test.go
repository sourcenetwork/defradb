// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package index

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestPatchRelation_OneToOne_AddsUniqueIndex(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
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
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						ID:     1,
						Name:   "Author__publishedID_ASC",
						Unique: true,
						Fields: []client.IndexedFieldDescription{
							{Name: "_publishedID"},
						},
					},
				},
			},
			&action.ListIndexes{
				CollectionID:    1,
				ExpectedIndexes: []client.IndexDescription{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestPatchRelation_MultipleOneToOne_AddsUniqueIndexesWithCorrectIDs(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Author {
						name: String
					}
					type Book {
						title: String
					}
					type Publisher {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Book/Fields/-", "value": {
							"Name": "author", "Kind": "Author", "RelationName": "book_author", "IsPrimary": true
						}},
						{ "op": "add", "path": "/Book/Fields/-", "value": {
							"Name": "_authorID", "Kind": 1, "RelationName": "book_author", "IsPrimary": true
						}},
						{ "op": "add", "path": "/Author/Fields/-", "value": {
							"Name": "book", "Kind": "Book", "RelationName": "book_author"
						}},
						{ "op": "add", "path": "/Book/Fields/-", "value": {
							"Name": "publisher", "Kind": "Publisher", "RelationName": "book_publisher", "IsPrimary": true
						}},
						{ "op": "add", "path": "/Book/Fields/-", "value": {
							"Name": "_publisherID", "Kind": 1, "RelationName": "book_publisher", "IsPrimary": true
						}},
						{ "op": "add", "path": "/Publisher/Fields/-", "value": {
							"Name": "book", "Kind": "Book", "RelationName": "book_publisher"
						}}
					]
				`,
			},
			&action.ListIndexes{
				CollectionID: 1,
				ExpectedIndexes: []client.IndexDescription{
					{
						ID:     1,
						Name:   "Book__authorID_ASC",
						Unique: true,
						Fields: []client.IndexedFieldDescription{
							{Name: "_authorID"},
						},
					},
					{
						ID:     2,
						Name:   "Book__publisherID_ASC",
						Unique: true,
						Fields: []client.IndexedFieldDescription{
							{Name: "_publisherID"},
						},
					},
				},
			},
			&action.ListIndexes{
				CollectionID:    0,
				ExpectedIndexes: []client.IndexDescription{},
			},
			&action.ListIndexes{
				CollectionID:    2,
				ExpectedIndexes: []client.IndexDescription{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestPatchRelation_OneToMany_DoesNotAddUniqueIndex(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
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
			&action.ListIndexes{
				CollectionID:    0,
				ExpectedIndexes: []client.IndexDescription{},
			},
			&action.ListIndexes{
				CollectionID:    1,
				ExpectedIndexes: []client.IndexDescription{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestPatchRelation_OneToOneWithVersionSwitching_IndexOnlyOnActiveVersion(t *testing.T) {
	const (
		authorV1 = "bafyreibvcavbxqwimz5vdxe5q5href63g3skc6ytg45hm4fqh6wsx57wmq"
		authorV2 = "bafyreihr72os6adcvjpsex4phzeefe6k32szyuqdgmyj7vfgvadxulyw5i"
	)

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
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
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						ID:     1,
						Name:   "Author__publishedID_ASC",
						Unique: true,
						Fields: []client.IndexedFieldDescription{
							{Name: "_publishedID"},
						},
					},
				},
			},
			testUtils.SetActiveCollectionVersion{
				VersionID: authorV1,
			},
			&action.ListIndexes{
				CollectionID:    0,
				ExpectedIndexes: []client.IndexDescription{},
			},
			testUtils.SetActiveCollectionVersion{
				VersionID: authorV2,
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						ID:     1,
						Name:   "Author__publishedID_ASC",
						Unique: true,
						Fields: []client.IndexedFieldDescription{
							{Name: "_publishedID"},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

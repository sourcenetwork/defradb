// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import (
	"context"
	"testing"

	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/client"
)

func TestSingleSimpleType(t *testing.T) {
	cases := []descriptionTestCase{
		{
			description: "Single simple type",
			sdl: `
			type User {
				name: String
				age: Int
				verified: Boolean
			}
			`,
			targetDescs: []client.CollectionDefinition{
				{
					Description: client.CollectionDescription{
						Name:    immutable.Some("User"),
						Indexes: []client.IndexDescription{},
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
							},
							{
								Name: "age",
							},
							{
								Name: "name",
							},
							{
								Name: "verified",
							},
						},
					},
					Schema: client.SchemaDescription{
						Name: "User",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "age",
								Kind: client.FieldKind_NILLABLE_INT,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "verified",
								Kind: client.FieldKind_NILLABLE_BOOL,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
		},
		{
			description: "Multiple simple types",
			sdl: `
			type User {
				name: String
				age: Int
				verified: Boolean
			}

			type Author {
				name: String
				publisher: String
				rating: Float
			}
			`,
			targetDescs: []client.CollectionDefinition{
				{
					Description: client.CollectionDescription{
						Name:    immutable.Some("User"),
						Indexes: []client.IndexDescription{},
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
							},
							{
								Name: "age",
							},
							{
								Name: "name",
							},
							{
								Name: "verified",
							},
						},
					},
					Schema: client.SchemaDescription{
						Name: "User",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "age",
								Kind: client.FieldKind_NILLABLE_INT,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "verified",
								Kind: client.FieldKind_NILLABLE_BOOL,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
				{
					Description: client.CollectionDescription{
						Name:    immutable.Some("Author"),
						Indexes: []client.IndexDescription{},
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
							},
							{
								Name: "name",
							},
							{
								Name: "publisher",
							},
							{
								Name: "rating",
							},
						},
					},
					Schema: client.SchemaDescription{
						Name: "Author",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "publisher",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "rating",
								Kind: client.FieldKind_NILLABLE_FLOAT,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
		},
		{
			description: "Multiple types with relations (one-to-one)",
			sdl: `
			type Book {
				name: String
				rating: Float
				author: Author
			}

			type Author {
				name: String
				age: Int
				published: Book @primary
			}
			`,
			targetDescs: []client.CollectionDefinition{
				{
					Description: client.CollectionDescription{
						Name:    immutable.Some("Book"),
						Indexes: []client.IndexDescription{},
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
							},
							{
								Name:         "author",
								Kind:         immutable.Some[client.FieldKind](client.ObjectKind("Author")),
								RelationName: immutable.Some("author_book"),
							},
							{
								Name:         "author_id",
								Kind:         immutable.Some[client.FieldKind](client.FieldKind_DocID),
								RelationName: immutable.Some("author_book"),
							},
							{
								Name: "name",
							},
							{
								Name: "rating",
							},
						},
					},
					Schema: client.SchemaDescription{
						Name: "Book",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "rating",
								Kind: client.FieldKind_NILLABLE_FLOAT,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
				{
					Description: client.CollectionDescription{
						Name:    immutable.Some("Author"),
						Indexes: []client.IndexDescription{},
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
							},
							{
								Name: "age",
							},
							{
								Name: "name",
							},
							{
								Name:         "published",
								Kind:         immutable.Some[client.FieldKind](client.ObjectKind("Book")),
								RelationName: immutable.Some("author_book"),
							},
							{
								Name:         "published_id",
								Kind:         immutable.Some[client.FieldKind](client.FieldKind_DocID),
								RelationName: immutable.Some("author_book"),
							},
						},
					},
					Schema: client.SchemaDescription{
						Name: "Author",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "age",
								Kind: client.FieldKind_NILLABLE_INT,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "published",
								Kind: client.ObjectKind("Book"),
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "published_id",
								Kind: client.FieldKind_DocID,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
		},
		{
			description: "Multiple types with relations (one-to-one)",
			sdl: `
			type Book {
				name: String
				rating: Float
				author: Author @relation(name:"book_authors")
			}

			type Author {
				name: String
				age: Int
				published: Book @relation(name:"book_authors") @primary
			}
			`,
			targetDescs: []client.CollectionDefinition{
				{
					Description: client.CollectionDescription{
						Name:    immutable.Some("Book"),
						Indexes: []client.IndexDescription{},
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
							},
							{
								Name:         "author",
								Kind:         immutable.Some[client.FieldKind](client.ObjectKind("Author")),
								RelationName: immutable.Some("book_authors"),
							},
							{
								Name:         "author_id",
								Kind:         immutable.Some[client.FieldKind](client.FieldKind_DocID),
								RelationName: immutable.Some("book_authors"),
							},
							{
								Name: "name",
							},
							{
								Name: "rating",
							},
						},
					},
					Schema: client.SchemaDescription{
						Name: "Book",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "rating",
								Kind: client.FieldKind_NILLABLE_FLOAT,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
				{
					Description: client.CollectionDescription{
						Name:    immutable.Some("Author"),
						Indexes: []client.IndexDescription{},
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
							},
							{
								Name: "age",
							},
							{
								Name: "name",
							},
							{
								Name:         "published",
								Kind:         immutable.Some[client.FieldKind](client.ObjectKind("Book")),
								RelationName: immutable.Some("book_authors"),
							},
							{
								Name:         "published_id",
								Kind:         immutable.Some[client.FieldKind](client.FieldKind_DocID),
								RelationName: immutable.Some("book_authors"),
							},
						},
					},
					Schema: client.SchemaDescription{
						Name: "Author",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "age",
								Kind: client.FieldKind_NILLABLE_INT,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "published",
								Kind: client.ObjectKind("Book"),
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "published_id",
								Kind: client.FieldKind_DocID,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
		},
		{
			description: "Multiple types with relations (one-to-one) with directive",
			sdl: `
			type Book {
				name: String
				rating: Float
				author: Author @primary
			}

			type Author {
				name: String
				age: Int
				published: Book
			}
			`,
			targetDescs: []client.CollectionDefinition{
				{
					Description: client.CollectionDescription{
						Name:    immutable.Some("Book"),
						Indexes: []client.IndexDescription{},
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
							},
							{
								Name:         "author",
								Kind:         immutable.Some[client.FieldKind](client.ObjectKind("Author")),
								RelationName: immutable.Some("author_book"),
							},
							{
								Name:         "author_id",
								Kind:         immutable.Some[client.FieldKind](client.FieldKind_DocID),
								RelationName: immutable.Some("author_book"),
							},
							{
								Name: "name",
							},
							{
								Name: "rating",
							},
						},
					},
					Schema: client.SchemaDescription{
						Name: "Book",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "author",
								Kind: client.ObjectKind("Author"),
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "author_id",
								Kind: client.FieldKind_DocID,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "rating",
								Kind: client.FieldKind_NILLABLE_FLOAT,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
				{
					Description: client.CollectionDescription{
						Name:    immutable.Some("Author"),
						Indexes: []client.IndexDescription{},
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
							},
							{
								Name: "age",
							},
							{
								Name: "name",
							},
							{
								Name:         "published",
								Kind:         immutable.Some[client.FieldKind](client.ObjectKind("Book")),
								RelationName: immutable.Some("author_book"),
							},
							{
								Name:         "published_id",
								Kind:         immutable.Some[client.FieldKind](client.FieldKind_DocID),
								RelationName: immutable.Some("author_book"),
							},
						},
					},
					Schema: client.SchemaDescription{
						Name: "Author",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "age",
								Kind: client.FieldKind_NILLABLE_INT,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
		},
		{
			description: "Multiple types with relations (one-to-many)",
			sdl: `
			type Book {
				name: String
				rating: Float
				author: Author
			}

			type Author {
				name: String
				age: Int
				published: [Book]
			}
			`,
			targetDescs: []client.CollectionDefinition{
				{
					Description: client.CollectionDescription{
						Name:    immutable.Some("Book"),
						Indexes: []client.IndexDescription{},
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
							},
							{
								Name:         "author",
								Kind:         immutable.Some[client.FieldKind](client.ObjectKind("Author")),
								RelationName: immutable.Some("author_book"),
							},
							{
								Name:         "author_id",
								Kind:         immutable.Some[client.FieldKind](client.FieldKind_DocID),
								RelationName: immutable.Some("author_book"),
							},
							{
								Name: "name",
							},
							{
								Name: "rating",
							},
						},
					},
					Schema: client.SchemaDescription{
						Name: "Book",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "rating",
								Kind: client.FieldKind_NILLABLE_FLOAT,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "author",
								Kind: client.ObjectKind("Author"),
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "author_id",
								Kind: client.FieldKind_DocID,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
				{
					Description: client.CollectionDescription{
						Name:    immutable.Some("Author"),
						Indexes: []client.IndexDescription{},
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
							},
							{
								Name: "age",
							},
							{
								Name: "name",
							},
							{
								Name:         "published",
								Kind:         immutable.Some[client.FieldKind](client.ObjectArrayKind("Book")),
								RelationName: immutable.Some("author_book"),
							},
						},
					},
					Schema: client.SchemaDescription{
						Name: "Author",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "age",
								Kind: client.FieldKind_NILLABLE_INT,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
		},
	}

	for _, test := range cases {
		runCreateDescriptionTest(t, test)
	}
}

func runCreateDescriptionTest(t *testing.T, testcase descriptionTestCase) {
	ctx := context.Background()

	descs, err := FromString(ctx, testcase.sdl)
	assert.NoError(t, err, testcase.description)
	assert.Equal(t, len(descs), len(testcase.targetDescs), testcase.description)

	for i, d := range descs {
		assert.Equal(t, testcase.targetDescs[i].Description, d.Description, testcase.description)
		assert.Equal(t, testcase.targetDescs[i].Schema, d.Schema, testcase.description)
	}
}

type descriptionTestCase struct {
	description string
	sdl         string
	targetDescs []client.CollectionDefinition
}

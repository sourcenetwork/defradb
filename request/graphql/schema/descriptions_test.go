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
			targetDescs: []client.CollectionDescription{
				{
					Name: "User",
					Schema: client.SchemaDescription{
						Name: "User",
						Fields: []client.FieldDescription{
							{
								Name: "_key",
								Kind: client.FieldKind_DocKey,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "age",
								Kind: client.FieldKind_INT,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "name",
								Kind: client.FieldKind_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "verified",
								Kind: client.FieldKind_BOOL,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
					Indexes: []client.IndexDescription{},
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
			targetDescs: []client.CollectionDescription{
				{
					Name: "User",
					Schema: client.SchemaDescription{
						Name: "User",
						Fields: []client.FieldDescription{
							{
								Name: "_key",
								Kind: client.FieldKind_DocKey,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "age",
								Kind: client.FieldKind_INT,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "name",
								Kind: client.FieldKind_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "verified",
								Kind: client.FieldKind_BOOL,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
					Indexes: []client.IndexDescription{},
				},
				{
					Name: "Author",
					Schema: client.SchemaDescription{
						Name: "Author",
						Fields: []client.FieldDescription{
							{
								Name: "_key",
								Kind: client.FieldKind_DocKey,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "name",
								Kind: client.FieldKind_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "publisher",
								Kind: client.FieldKind_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "rating",
								Kind: client.FieldKind_FLOAT,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
					Indexes: []client.IndexDescription{},
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
				published: Book
			}
			`,
			targetDescs: []client.CollectionDescription{
				{
					Name: "Book",
					Schema: client.SchemaDescription{
						Name: "Book",
						Fields: []client.FieldDescription{
							{
								Name: "_key",
								Kind: client.FieldKind_DocKey,
								Typ:  client.NONE_CRDT,
							},
							{
								Name:         "author",
								RelationName: "author_book",
								Kind:         client.FieldKind_FOREIGN_OBJECT,
								Typ:          client.NONE_CRDT,
								Schema:       "Author",
								RelationType: client.Relation_Type_ONE | client.Relation_Type_ONEONE,
							},
							{
								Name:         "author_id",
								Kind:         client.FieldKind_DocKey,
								Typ:          client.LWW_REGISTER,
								RelationType: client.Relation_Type_INTERNAL_ID,
							},
							{
								Name: "name",
								Kind: client.FieldKind_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "rating",
								Kind: client.FieldKind_FLOAT,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
					Indexes: []client.IndexDescription{},
				},
				{
					Name: "Author",
					Schema: client.SchemaDescription{
						Name: "Author",
						Fields: []client.FieldDescription{
							{
								Name: "_key",
								Kind: client.FieldKind_DocKey,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "age",
								Kind: client.FieldKind_INT,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "name",
								Kind: client.FieldKind_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name:         "published",
								RelationName: "author_book",
								Kind:         client.FieldKind_FOREIGN_OBJECT,
								Typ:          client.NONE_CRDT,
								Schema:       "Book",
								RelationType: client.Relation_Type_ONE | client.Relation_Type_ONEONE | client.Relation_Type_Primary,
							},
							{
								Name:         "published_id",
								Kind:         client.FieldKind_DocKey,
								Typ:          client.LWW_REGISTER,
								RelationType: client.Relation_Type_INTERNAL_ID,
							},
						},
					},
					Indexes: []client.IndexDescription{},
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
			targetDescs: []client.CollectionDescription{
				{
					Name: "User",
					Schema: client.SchemaDescription{
						Name: "User",
						Fields: []client.FieldDescription{
							{
								Name: "_key",
								Kind: client.FieldKind_DocKey,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "age",
								Kind: client.FieldKind_INT,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "name",
								Kind: client.FieldKind_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "verified",
								Kind: client.FieldKind_BOOL,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
					Indexes: []client.IndexDescription{},
				},
				{
					Name: "Author",
					Schema: client.SchemaDescription{
						Name: "Author",
						Fields: []client.FieldDescription{
							{
								Name: "_key",
								Kind: client.FieldKind_DocKey,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "name",
								Kind: client.FieldKind_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "publisher",
								Kind: client.FieldKind_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "rating",
								Kind: client.FieldKind_FLOAT,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
					Indexes: []client.IndexDescription{},
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
				published: Book @relation(name:"book_authors")
			}
			`,
			targetDescs: []client.CollectionDescription{
				{
					Name: "Book",
					Schema: client.SchemaDescription{
						Name: "Book",
						Fields: []client.FieldDescription{
							{
								Name: "_key",
								Kind: client.FieldKind_DocKey,
								Typ:  client.NONE_CRDT,
							},
							{
								Name:         "author",
								RelationName: "book_authors",
								Kind:         client.FieldKind_FOREIGN_OBJECT,
								Typ:          client.NONE_CRDT,
								Schema:       "Author",
								RelationType: client.Relation_Type_ONE | client.Relation_Type_ONEONE,
							},
							{
								Name:         "author_id",
								Kind:         client.FieldKind_DocKey,
								Typ:          client.LWW_REGISTER,
								RelationType: client.Relation_Type_INTERNAL_ID,
							},
							{
								Name: "name",
								Kind: client.FieldKind_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "rating",
								Kind: client.FieldKind_FLOAT,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
					Indexes: []client.IndexDescription{},
				},
				{
					Name: "Author",
					Schema: client.SchemaDescription{
						Name: "Author",
						Fields: []client.FieldDescription{
							{
								Name: "_key",
								Kind: client.FieldKind_DocKey,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "age",
								Kind: client.FieldKind_INT,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "name",
								Kind: client.FieldKind_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name:         "published",
								RelationName: "book_authors",
								Kind:         client.FieldKind_FOREIGN_OBJECT,
								Typ:          client.NONE_CRDT,
								Schema:       "Book",
								RelationType: client.Relation_Type_ONE | client.Relation_Type_ONEONE | client.Relation_Type_Primary,
							},
							{
								Name:         "published_id",
								Kind:         client.FieldKind_DocKey,
								Typ:          client.LWW_REGISTER,
								RelationType: client.Relation_Type_INTERNAL_ID,
							},
						},
					},
					Indexes: []client.IndexDescription{},
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
			targetDescs: []client.CollectionDescription{
				{
					Name: "Book",
					Schema: client.SchemaDescription{
						Name: "Book",
						Fields: []client.FieldDescription{
							{
								Name: "_key",
								Kind: client.FieldKind_DocKey,
								Typ:  client.NONE_CRDT,
							},
							{
								Name:         "author",
								RelationName: "author_book",
								Kind:         client.FieldKind_FOREIGN_OBJECT,
								Typ:          client.NONE_CRDT,
								Schema:       "Author",
								RelationType: client.Relation_Type_ONE | client.Relation_Type_ONEONE | client.Relation_Type_Primary,
							},
							{
								Name:         "author_id",
								Kind:         client.FieldKind_DocKey,
								Typ:          client.LWW_REGISTER,
								RelationType: client.Relation_Type_INTERNAL_ID,
							},
							{
								Name: "name",
								Kind: client.FieldKind_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "rating",
								Kind: client.FieldKind_FLOAT,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
					Indexes: []client.IndexDescription{},
				},
				{
					Name: "Author",
					Schema: client.SchemaDescription{
						Name: "Author",
						Fields: []client.FieldDescription{
							{
								Name: "_key",
								Kind: client.FieldKind_DocKey,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "age",
								Kind: client.FieldKind_INT,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "name",
								Kind: client.FieldKind_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name:         "published",
								RelationName: "author_book",
								Kind:         client.FieldKind_FOREIGN_OBJECT,
								Typ:          client.NONE_CRDT,
								Schema:       "Book",
								RelationType: client.Relation_Type_ONE | client.Relation_Type_ONEONE,
							},
							{
								Name:         "published_id",
								Kind:         client.FieldKind_DocKey,
								Typ:          client.LWW_REGISTER,
								RelationType: client.Relation_Type_INTERNAL_ID,
							},
						},
					},
					Indexes: []client.IndexDescription{},
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
			targetDescs: []client.CollectionDescription{
				{
					Name: "Book",
					Schema: client.SchemaDescription{
						Name: "Book",
						Fields: []client.FieldDescription{
							{
								Name: "_key",
								Kind: client.FieldKind_DocKey,
								Typ:  client.NONE_CRDT,
							},
							{
								Name:         "author",
								RelationName: "author_book",
								Kind:         client.FieldKind_FOREIGN_OBJECT,
								Typ:          client.NONE_CRDT,
								Schema:       "Author",
								RelationType: client.Relation_Type_ONE | client.Relation_Type_ONEMANY | client.Relation_Type_Primary,
							},
							{
								Name:         "author_id",
								Kind:         client.FieldKind_DocKey,
								Typ:          client.LWW_REGISTER,
								RelationType: client.Relation_Type_INTERNAL_ID,
							},
							{
								Name: "name",
								Kind: client.FieldKind_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "rating",
								Kind: client.FieldKind_FLOAT,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
					Indexes: []client.IndexDescription{},
				},
				{
					Name: "Author",
					Schema: client.SchemaDescription{
						Name: "Author",
						Fields: []client.FieldDescription{
							{
								Name: "_key",
								Kind: client.FieldKind_DocKey,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "age",
								Kind: client.FieldKind_INT,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "name",
								Kind: client.FieldKind_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name:         "published",
								RelationName: "author_book",
								Kind:         client.FieldKind_FOREIGN_OBJECT_ARRAY,
								Typ:          client.NONE_CRDT,
								Schema:       "Book",
								RelationType: client.Relation_Type_MANY | client.Relation_Type_ONEMANY,
							},
						},
					},
					Indexes: []client.IndexDescription{},
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
		assert.Equal(t, testcase.targetDescs[i], d, testcase.description)
	}
}

type descriptionTestCase struct {
	description string
	sdl         string
	targetDescs []client.CollectionDescription
}

package schema

import (
	"testing"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"

	"github.com/stretchr/testify/assert"
)

var testDefaultIndex = []base.IndexDescription{
	base.IndexDescription{
		Name:    "primary",
		ID:      uint32(0),
		Primary: true,
		Unique:  true,
	},
}

func TestSingleSimpleType(t *testing.T) {
	cases := []descriptionTestCase{
		{
			description: "Single simple type",
			sdl: `
			type user {
				name: String
				age: Int
				verified: Boolean
			}
			`,
			targetDescs: []base.CollectionDescription{
				base.CollectionDescription{
					Name: "user",
					Schema: base.SchemaDescription{
						Name: "user",
						Fields: []base.FieldDescription{
							{
								Name: "_key",
								Kind: base.FieldKind_DocKey,
								Typ:  core.NONE_CRDT,
							},
							{
								Name: "age",
								Kind: base.FieldKind_INT,
								Typ:  core.LWW_REGISTER,
							},
							{
								Name: "name",
								Kind: base.FieldKind_STRING,
								Typ:  core.LWW_REGISTER,
							},
							{
								Name: "verified",
								Kind: base.FieldKind_BOOL,
								Typ:  core.LWW_REGISTER,
							},
						},
					},
					Indexes: testDefaultIndex,
				},
			},
		},
		{
			description: "Multiple simple types",
			sdl: `
			type user {
				name: String
				age: Int
				verified: Boolean
			}

			type author {
				name: String
				publisher: String
				rating: Float
			}
			`,
			targetDescs: []base.CollectionDescription{
				base.CollectionDescription{
					Name: "user",
					Schema: base.SchemaDescription{
						Name: "user",
						Fields: []base.FieldDescription{
							{
								Name: "_key",
								Kind: base.FieldKind_DocKey,
								Typ:  core.NONE_CRDT,
							},
							{
								Name: "age",
								Kind: base.FieldKind_INT,
								Typ:  core.LWW_REGISTER,
							},
							{
								Name: "name",
								Kind: base.FieldKind_STRING,
								Typ:  core.LWW_REGISTER,
							},
							{
								Name: "verified",
								Kind: base.FieldKind_BOOL,
								Typ:  core.LWW_REGISTER,
							},
						},
					},
					Indexes: testDefaultIndex,
				},
				base.CollectionDescription{
					Name: "author",
					Schema: base.SchemaDescription{
						Name: "author",
						Fields: []base.FieldDescription{
							{
								Name: "_key",
								Kind: base.FieldKind_DocKey,
								Typ:  core.NONE_CRDT,
							},
							{
								Name: "name",
								Kind: base.FieldKind_STRING,
								Typ:  core.LWW_REGISTER,
							},
							{
								Name: "publisher",
								Kind: base.FieldKind_STRING,
								Typ:  core.LWW_REGISTER,
							},
							{
								Name: "rating",
								Kind: base.FieldKind_FLOAT,
								Typ:  core.LWW_REGISTER,
							},
						},
					},
					Indexes: testDefaultIndex,
				},
			},
		},
		{
			description: "Multiple types with relations (one-to-one)",
			sdl: `
			type book {
				name: String
				rating: Float
				author: author
			}

			type author {
				name: String
				age: Int
				published: book
			}
			`,
			targetDescs: []base.CollectionDescription{
				base.CollectionDescription{
					Name: "book",
					Schema: base.SchemaDescription{
						Name: "book",
						Fields: []base.FieldDescription{
							{
								Name: "_key",
								Kind: base.FieldKind_DocKey,
								Typ:  core.NONE_CRDT,
							},
							{
								Name:   "author",
								Kind:   base.FieldKind_FOREIGN_OBJECT,
								Typ:    core.NONE_CRDT,
								Schema: "author",
								Meta:   base.Meta_Relation_ONE | base.Meta_Relation_ONEONE,
							},
							{
								Name: "author_id",
								Kind: base.FieldKind_DocKey,
								Typ:  core.LWW_REGISTER,
							},
							{
								Name: "name",
								Kind: base.FieldKind_STRING,
								Typ:  core.LWW_REGISTER,
							},
							{
								Name: "rating",
								Kind: base.FieldKind_FLOAT,
								Typ:  core.LWW_REGISTER,
							},
						},
					},
					Indexes: testDefaultIndex,
				},
				base.CollectionDescription{
					Name: "author",
					Schema: base.SchemaDescription{
						Name: "author",
						Fields: []base.FieldDescription{
							{
								Name: "_key",
								Kind: base.FieldKind_DocKey,
								Typ:  core.NONE_CRDT,
							},
							{
								Name: "age",
								Kind: base.FieldKind_INT,
								Typ:  core.LWW_REGISTER,
							},
							{
								Name: "name",
								Kind: base.FieldKind_STRING,
								Typ:  core.LWW_REGISTER,
							},
							{
								Name:   "published",
								Kind:   base.FieldKind_FOREIGN_OBJECT,
								Typ:    core.NONE_CRDT,
								Schema: "book",
								Meta:   base.Meta_Relation_ONE | base.Meta_Relation_ONEONE | base.Meta_Relation_Primary,
							},
							{
								Name: "published_id",
								Kind: base.FieldKind_DocKey,
								Typ:  core.LWW_REGISTER,
							},
						},
					},
					Indexes: testDefaultIndex,
				},
			},
		},
		{
			description: "Multiple types with relations (one-to-one)",
			sdl: `
			type book {
				name: String
				rating: Float
				author: author
			}

			type author {
				name: String
				age: Int
				published: [book]
			}
			`,
			targetDescs: []base.CollectionDescription{
				base.CollectionDescription{
					Name: "book",
					Schema: base.SchemaDescription{
						Name: "book",
						Fields: []base.FieldDescription{
							{
								Name: "_key",
								Kind: base.FieldKind_DocKey,
								Typ:  core.NONE_CRDT,
							},
							{
								Name:   "author",
								Kind:   base.FieldKind_FOREIGN_OBJECT,
								Typ:    core.NONE_CRDT,
								Schema: "author",
								Meta:   base.Meta_Relation_ONE | base.Meta_Relation_ONEMANY | base.Meta_Relation_Primary,
							},
							{
								Name: "author_id",
								Kind: base.FieldKind_DocKey,
								Typ:  core.LWW_REGISTER,
							},
							{
								Name: "name",
								Kind: base.FieldKind_STRING,
								Typ:  core.LWW_REGISTER,
							},
							{
								Name: "rating",
								Kind: base.FieldKind_FLOAT,
								Typ:  core.LWW_REGISTER,
							},
						},
					},
					Indexes: testDefaultIndex,
				},
				base.CollectionDescription{
					Name: "author",
					Schema: base.SchemaDescription{
						Name: "author",
						Fields: []base.FieldDescription{
							{
								Name: "_key",
								Kind: base.FieldKind_DocKey,
								Typ:  core.NONE_CRDT,
							},
							{
								Name: "age",
								Kind: base.FieldKind_INT,
								Typ:  core.LWW_REGISTER,
							},
							{
								Name: "name",
								Kind: base.FieldKind_STRING,
								Typ:  core.LWW_REGISTER,
							},
							{
								Name:   "published",
								Kind:   base.FieldKind_FOREIGN_OBJECT_ARRAY,
								Typ:    core.NONE_CRDT,
								Schema: "book",
								Meta:   base.Meta_Relation_MANY | base.Meta_Relation_ONEMANY,
							},
						},
					},
					Indexes: testDefaultIndex,
				},
			},
		},
	}

	for _, test := range cases {
		runCreateDescriptionTest(t, test)
	}
}

func runCreateDescriptionTest(t *testing.T, testcase descriptionTestCase) {
	sm, err := NewSchemaManager()
	assert.NoError(t, err, testcase.description)

	types, err := sm.Generator.FromSDL(testcase.sdl)
	assert.NoError(t, err, testcase.description)

	assert.Len(t, types, len(testcase.targetDescs), testcase.description)

	descs, err := sm.Generator.CreateDescriptions(types)
	assert.NoError(t, err, testcase.description)
	assert.Equal(t, len(descs), len(testcase.targetDescs), testcase.description)

	for i, d := range descs {
		// fmt.Println(d.Schema.Fields)
		assert.Equal(t, testcase.targetDescs[i], d, testcase.description)
	}
}

type descriptionTestCase struct {
	description string
	sdl         string
	targetDescs []base.CollectionDescription
}

// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package db

// var userCollectionGQLSchema = (`
// type users {
// 	Name: String
// 	Age: Int
// }
// `)

// // func newQueryableDB()

// func TestSimpleCollectionQuery(t *testing.T) {
// 	db, err := newMemoryDB()
// 	assert.NoError(t, err)

// 	desc := newTestCollectionDescription()
// 	col, err := db.CreateCollection(desc)
// 	assert.NoError(t, err)

// 	executor, err := planner.NewQueryExecutor()
// 	assert.NoError(t, err)

// 	err = executor.Generator.FromSDL(userCollectionGQLSchema)
// 	assert.NoError(t, err)

// 	doc1, err := document.NewFromJSON([]byte(`{
// 		"Name": "John",
// 		"Age": 21
// 	}`))

// 	assert.NoError(t, err)
// 	err = col.Save(doc1)
// 	assert.NoError(t, err)

// 	txn, err := db.NewTxn(true)
// 	assert.NoError(t, err)

// 	// obj := executor.SchemaManager.Schema().TypeMap()["users"].(*gql.Object)
// 	// obj.Fields()
// 	// spew.Dump(obj.Fields())

// 	var userQuery = (`
// 	query {
// 		users {
// 			Name
// 			Age
// 		}
// 	}`)

// 	docs, err := executor.ExecQuery(txn, userQuery)
// 	assert.NoError(t, err)

// 	fmt.Println(docs)
// 	assert.True(t, len(docs) == 1)
// }

// func TestSimpleCollectionQueryWithFilter(t *testing.T) {
// 	db, err := newMemoryDB()
// 	assert.NoError(t, err)

// 	desc := newTestCollectionDescription()
// 	col, err := db.CreateCollection(desc)
// 	assert.NoError(t, err)

// 	executor, err := planner.NewQueryExecutor()
// 	assert.NoError(t, err)

// 	err = executor.Generator.FromSDL(userCollectionGQLSchema)
// 	assert.NoError(t, err)

// 	doc1, err := document.NewFromJSON([]byte(`{
// 		"Name": "John",
// 		"Age": 21
// 	}`))

// 	assert.NoError(t, err)
// 	err = col.Save(doc1)
// 	assert.NoError(t, err)

// 	txn, err := db.NewTxn(true)
// 	assert.NoError(t, err)

// 	// obj := executor.SchemaManager.Schema().TypeMap()["users"].(*gql.Object)
// 	// obj.Fields()
// 	// spew.Dump(obj.Fields())

// 	var userQuery = (`
// 	query {
// 		users(filter: {Name: {_eq: "John"}}) {
// 			Name
// 			Age
// 		}
// 	}`)

// 	docs, err := executor.ExecQuery(txn, userQuery)
// 	assert.NoError(t, err)

// 	// fmt.Println(docs)
// 	assert.Len(t, docs, 1)

// 	assert.Equal(t, map[string]interface{}{
// 		"Name": "John",
// 		"Age":  uint64(21),
// 	}, docs[0])
// }

// func newTestQueryCollectionDescription1() base.CollectionDescription {
// 	return base.CollectionDescription{
// 		Name: "users",
// 		ID:   uint32(1),
// 		Schema: base.SchemaDescription{
// 			ID:       uint32(1),
// 			FieldIDs: []uint32{1, 2, 3, 5},
// 			Fields: []base.FieldDescription{
// 				base.FieldDescription{
// 					Name: "_key",
// 					ID:   base.FieldID(1),
// 					Kind: base.FieldKind_DocKey,
// 				},
// 				base.FieldDescription{
// 					Name: "Name",
// 					ID:   base.FieldID(2),
// 					Kind: base.FieldKind_STRING,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 				base.FieldDescription{
// 					Name: "Age",
// 					ID:   base.FieldID(3),
// 					Kind: base.FieldKind_INT,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 				base.FieldDescription{
// 					Name: "Verified",
// 					ID:   base.FieldID(4),
// 					Kind: base.FieldKind_BOOL,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 			},
// 		},
// 		Indexes: []base.IndexDescription{
// 			base.IndexDescription{
// 				Name:    "primary",
// 				ID:      uint32(0),
// 				Primary: true,
// 				Unique:  true,
// 			},
// 		},
// 	}
// }

// func newTestQueryCollectionDescription2() base.CollectionDescription {
// 	return base.CollectionDescription{
// 		Name: "book",
// 		ID:   uint32(2),
// 		Schema: base.SchemaDescription{
// 			ID:       uint32(2),
// 			FieldIDs: []uint32{1, 2, 3, 4, 5},
// 			Fields: []base.FieldDescription{
// 				base.FieldDescription{
// 					Name: "_key",
// 					ID:   base.FieldID(1),
// 					Kind: base.FieldKind_DocKey,
// 				},
// 				base.FieldDescription{
// 					Name: "name",
// 					ID:   base.FieldID(2),
// 					Kind: base.FieldKind_STRING,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 				base.FieldDescription{
// 					Name: "rating",
// 					ID:   base.FieldID(3),
// 					Kind: base.FieldKind_FLOAT,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 				base.FieldDescription{
// 					Name:   "author",
// 					ID:     base.FieldID(5),
// 					Kind:   base.FieldKind_FOREIGN_OBJECT,
// 					Schema: "author",
// 					Typ:    core.NONE_CRDT,
// 					Meta:   base.Meta_Relation_ONE | base.Meta_Relation_ONEONE | base.Meta_Relation_Primary,
// 				},
// 				base.FieldDescription{
// 					Name: "author_id",
// 					ID:   base.FieldID(6),
// 					Kind: base.FieldKind_DocKey,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 			},
// 		},
// 		Indexes: []base.IndexDescription{
// 			base.IndexDescription{
// 				Name:    "primary",
// 				ID:      uint32(0),
// 				Primary: true,
// 				Unique:  true,
// 			},
// 		},
// 	}
// }

// func newTestQueryCollectionDescription3() base.CollectionDescription {
// 	return base.CollectionDescription{
// 		Name: "author",
// 		ID:   uint32(3),
// 		Schema: base.SchemaDescription{
// 			ID:       uint32(3),
// 			Name:     "author",
// 			FieldIDs: []uint32{1, 2, 3, 4, 5, 6},
// 			Fields: []base.FieldDescription{
// 				base.FieldDescription{
// 					Name: "_key",
// 					ID:   base.FieldID(1),
// 					Kind: base.FieldKind_DocKey,
// 				},
// 				base.FieldDescription{
// 					Name: "name",
// 					ID:   base.FieldID(2),
// 					Kind: base.FieldKind_STRING,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 				base.FieldDescription{
// 					Name: "age",
// 					ID:   base.FieldID(3),
// 					Kind: base.FieldKind_INT,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 				base.FieldDescription{
// 					Name: "verified",
// 					ID:   base.FieldID(4),
// 					Kind: base.FieldKind_BOOL,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 				base.FieldDescription{
// 					Name:   "published",
// 					ID:     base.FieldID(5),
// 					Kind:   base.FieldKind_FOREIGN_OBJECT,
// 					Schema: "book",
// 					Typ:    core.NONE_CRDT,
// 					Meta:   base.Meta_Relation_ONE,
// 				},
// 				base.FieldDescription{
// 					Name: "published_id",
// 					ID:   base.FieldID(6),
// 					Kind: base.FieldKind_DocKey,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 			},
// 		},
// 		Indexes: []base.IndexDescription{
// 			base.IndexDescription{
// 				Name:    "primary",
// 				ID:      uint32(0),
// 				Primary: true,
// 				Unique:  true,
// 			},
// 		},
// 	}
// }

// func newTestQueryCollectionDescription4() base.CollectionDescription {
// 	return base.CollectionDescription{
// 		Name: "book",
// 		ID:   uint32(2),
// 		Schema: base.SchemaDescription{
// 			ID:       uint32(2),
// 			FieldIDs: []uint32{1, 2, 3, 4, 5},
// 			Fields: []base.FieldDescription{
// 				base.FieldDescription{
// 					Name: "_key",
// 					ID:   base.FieldID(1),
// 					Kind: base.FieldKind_DocKey,
// 				},
// 				base.FieldDescription{
// 					Name: "name",
// 					ID:   base.FieldID(2),
// 					Kind: base.FieldKind_STRING,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 				base.FieldDescription{
// 					Name: "rating",
// 					ID:   base.FieldID(3),
// 					Kind: base.FieldKind_FLOAT,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 				base.FieldDescription{
// 					Name:   "author",
// 					ID:     base.FieldID(5),
// 					Kind:   base.FieldKind_FOREIGN_OBJECT,
// 					Schema: "author",
// 					Typ:    core.NONE_CRDT,
// 					Meta:   base.Meta_Relation_ONE | base.Meta_Relation_ONEMANY | base.Meta_Relation_Primary,
// 				},
// 				base.FieldDescription{
// 					Name: "author_id",
// 					ID:   base.FieldID(6),
// 					Kind: base.FieldKind_DocKey,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 			},
// 		},
// 		Indexes: []base.IndexDescription{
// 			base.IndexDescription{
// 				Name:    "primary",
// 				ID:      uint32(0),
// 				Primary: true,
// 				Unique:  true,
// 			},
// 		},
// 	}
// }

// func newTestQueryCollectionDescription5() base.CollectionDescription {
// 	return base.CollectionDescription{
// 		Name: "author",
// 		ID:   uint32(3),
// 		Schema: base.SchemaDescription{
// 			ID:       uint32(3),
// 			Name:     "author",
// 			FieldIDs: []uint32{1, 2, 3, 4, 5},
// 			Fields: []base.FieldDescription{
// 				base.FieldDescription{
// 					Name: "_key",
// 					ID:   base.FieldID(1),
// 					Kind: base.FieldKind_DocKey,
// 				},
// 				base.FieldDescription{
// 					Name: "name",
// 					ID:   base.FieldID(2),
// 					Kind: base.FieldKind_STRING,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 				base.FieldDescription{
// 					Name: "age",
// 					ID:   base.FieldID(3),
// 					Kind: base.FieldKind_INT,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 				base.FieldDescription{
// 					Name: "verified",
// 					ID:   base.FieldID(4),
// 					Kind: base.FieldKind_BOOL,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 				base.FieldDescription{
// 					Name:   "published",
// 					ID:     base.FieldID(5),
// 					Kind:   base.FieldKind_FOREIGN_OBJECT_ARRAY,
// 					Schema: "book",
// 					Typ:    core.NONE_CRDT,
// 					Meta:   base.Meta_Relation_MANY | base.Meta_Relation_ONEMANY,
// 				},
// 			},
// 		},
// 		Indexes: []base.IndexDescription{
// 			base.IndexDescription{
// 				Name:    "primary",
// 				ID:      uint32(0),
// 				Primary: true,
// 				Unique:  true,
// 			},
// 		},
// 	}
// }

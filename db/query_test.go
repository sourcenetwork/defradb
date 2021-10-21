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

import (
	"fmt"
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/document"
	"github.com/stretchr/testify/assert"
)

type queryTestCase struct {
	description string
	query       string
	// docs is a map from Collection Index, to a list
	// of docs in stringified JSON format
	docs map[int][]string
	// updates is a map from document index, to a list
	// of changes in strinigied JSON format
	updates map[int][]string
	results []map[string]interface{}
}

func runQueryTestCase(t *testing.T, db *DB, collections []client.Collection, test queryTestCase) {
	// insert docs
	for cid, docs := range test.docs {
		for i, docStr := range docs {
			doc, err := document.NewFromJSON([]byte(docStr))
			assert.NoError(t, err, test.description)
			err = collections[cid].Save(doc)
			assert.NoError(t, err, test.description)

			// check for updates
			updates, ok := test.updates[i]
			if ok {
				for _, u := range updates {
					err = doc.SetWithJSON([]byte(u))
					assert.NoError(t, err, test.description)
					err = collections[cid].Save(doc)
					assert.NoError(t, err, test.description)
				}
			}
		}
	}

	// exec query
	txn, err := db.NewTxn(true)
	assert.NoError(t, err, test.description)
	results, err := db.queryExecutor.ExecQuery(db, txn, test.query)
	assert.NoError(t, err, test.description)

	fmt.Println(test.description)
	fmt.Println(results)
	fmt.Println("--------------")
	fmt.Println("")

	// compare results
	assert.Equal(t, len(test.results), len(results), test.description)
	for i, result := range results {
		assert.Equal(t, test.results[i], result, test.description)
	}
}

func TestQueryAllCommitsSingleDAG(t *testing.T) {
	var userCollectionGQLSchema = (`
	type users {
		Name: String
		Age: Int
		Verified: Boolean
	}
	`)

	tests := []queryTestCase{
		{
			description: "Simple latest commits query",
			query: `query {
						allCommits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
							cid
							links {
								cid
								name
							}
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			results: []map[string]interface{}{
				{
					"cid": "QmaXdKKsc5GRWXtMytZj4PEf5hFgFxjZaKToQpDY8cAocV",
					"links": []map[string]interface{}{
						{
							"cid":  "QmPaY2DNmd7LtRDpReswc5UTGoU5Q32Py1aEVG7Shq6Np1",
							"name": "Age",
						},
						{
							"cid":  "Qmag2zKKGGQwVSss9pQn3hjTu9opdF5mkJXUR9rt2A651h",
							"name": "Name",
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		db, err := newMemoryDB()
		assert.NoError(t, err)

		err = db.AddSchema(userCollectionGQLSchema)
		assert.NoError(t, err)

		// desc := newTestQueryCollectionDescription1()
		col, err := db.GetCollection("users")
		assert.NoError(t, err)

		runQueryTestCase(t, db, []client.Collection{col}, test)
	}

}

func TestQueryAllCommitsMultipleDAG(t *testing.T) {
	var userCollectionGQLSchema = (`
	type users {
		Name: String
		Age: Int
		Verified: Boolean
	}
	`)

	tests := []queryTestCase{
		{
			description: "Simple latest commits query",
			query: `query {
						allCommits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
							cid
							height
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			updates: map[int][]string{
				0: []string{
					(`{"Age": 22}`), // update to change age to 22 on document 0
				},
			},
			results: []map[string]interface{}{
				{
					"cid":    "QmQQgYgC3PLFCTwsSgMHHFvFbPEeWDKkbsnvYJwuLP3R8t",
					"height": int64(2),
				},
				{
					"cid":    "QmaXdKKsc5GRWXtMytZj4PEf5hFgFxjZaKToQpDY8cAocV",
					"height": int64(1),
				},
			},
		},
	}

	for _, test := range tests {
		db, err := newMemoryDB()
		assert.NoError(t, err)

		err = db.AddSchema(userCollectionGQLSchema)
		assert.NoError(t, err)

		// desc := newTestQueryCollectionDescription1()
		col, err := db.GetCollection("users")
		assert.NoError(t, err)

		runQueryTestCase(t, db, []client.Collection{col}, test)
	}

}

func TestQueryEmbeddedLatestCommit(t *testing.T) {
	var userCollectionGQLSchema = (`
	type users {
		Name: String
		Age: Int
		Verified: Boolean
	}
	`)

	tests := []queryTestCase{
		{
			description: "Embedded latest commits query within object query",
			query: `query {
						users {
							Name
							Age
							_version {
								cid
								links {
									cid
									name
								}
							}
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			results: []map[string]interface{}{
				{
					"Name": "John",
					"Age":  uint64(21),
					"_version": []map[string]interface{}{
						{
							"cid": "QmaXdKKsc5GRWXtMytZj4PEf5hFgFxjZaKToQpDY8cAocV",
							"links": []map[string]interface{}{
								{
									"cid":  "QmPaY2DNmd7LtRDpReswc5UTGoU5Q32Py1aEVG7Shq6Np1",
									"name": "Age",
								},
								{
									"cid":  "Qmag2zKKGGQwVSss9pQn3hjTu9opdF5mkJXUR9rt2A651h",
									"name": "Name",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		db, err := newMemoryDB()
		assert.NoError(t, err)

		err = db.AddSchema(userCollectionGQLSchema)
		assert.NoError(t, err)

		// desc := newTestQueryCollectionDescription1()
		col, err := db.GetCollection("users")
		assert.NoError(t, err)

		runQueryTestCase(t, db, []client.Collection{col}, test)
	}

}

func TestQueryOneCommit(t *testing.T) {
	var userCollectionGQLSchema = (`
	type users {
		Name: String
		Age: Int
		Verified: Boolean
	}
	`)

	tests := []queryTestCase{
		{
			description: "query for a single block by CID",
			query: `query {
						commit(cid: "QmaXdKKsc5GRWXtMytZj4PEf5hFgFxjZaKToQpDY8cAocV") {
							cid
							height
							delta
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			results: []map[string]interface{}{
				{
					"cid":    "QmaXdKKsc5GRWXtMytZj4PEf5hFgFxjZaKToQpDY8cAocV",
					"height": int64(1),
					// cbor encoded delta
					"delta": []uint8{0xa2, 0x63, 0x41, 0x67, 0x65, 0x15, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x64, 0x4a, 0x6f, 0x68, 0x6e},
				},
			},
		},
	}

	for _, test := range tests {
		db, err := newMemoryDB()
		assert.NoError(t, err)

		err = db.AddSchema(userCollectionGQLSchema)
		assert.NoError(t, err)

		// desc := newTestQueryCollectionDescription1()
		col, err := db.GetCollection("users")
		assert.NoError(t, err)

		runQueryTestCase(t, db, []client.Collection{col}, test)
	}

}

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

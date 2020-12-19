package db

import (
	"fmt"
	"testing"

	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/query/graphql/planner"
	"github.com/stretchr/testify/assert"
)

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

type queryTestCase struct {
	description string
	query       string
	docs        []string
	results     []map[string]interface{}
}

func TestQuerySimple(t *testing.T) {
	var userCollectionGQLSchema = (`
	type users {
		Name: String
		Age: Int
	}
	`)
	db, err := newMemoryDB()
	assert.NoError(t, err)

	desc := newTestCollectionDescription()
	col, err := db.CreateCollection(desc)
	assert.NoError(t, err)

	executor, err := planner.NewQueryExecutor()
	assert.NoError(t, err)

	db.queryExecutor = executor

	err = executor.Generator.FromSDL(userCollectionGQLSchema)
	assert.NoError(t, err)

	tests := []queryTestCase{
		{
			description: "Simple query with no filter",
			query: `query {
						users {
							Name
							Age
						}
					}`,
			docs: []string{
				(`{
					"Name": "John",
					"Age": 21
				}`),
			},
			results: []map[string]interface{}{
				{
					"Name": "John",
					"Age":  uint64(21),
				},
			},
		},
		{
			description: "Simple query with basic filter",
			query: `query {
						users(filter: {Name: {_eq: "John"}}) {
							Name
							Age
						}
					}`,
			docs: []string{
				(`{
					"Name": "John",
					"Age": 21
				}`),
			},
			results: []map[string]interface{}{
				{
					"Name": "John",
					"Age":  uint64(21),
				},
			},
		},
		{
			description: "Simple query with basic filter(name), no results",
			query: `query {
						users(filter: {Name: {_eq: "Bob"}}) {
							Name
							Age
						}
					}`,
			docs: []string{
				(`{
					"Name": "John",
					"Age": 21
				}`),
			},
			results: []map[string]interface{}{},
		},
		{
			description: "Simple query with basic filter(age)",
			query: `query {
						users(filter: {Age: {_eq: 21}}) {
							Name
							Age
						}
					}`,
			docs: []string{
				(`{
					"Name": "John",
					"Age": 21
				}`),
			},
			results: []map[string]interface{}{
				{
					"Name": "John",
					"Age":  uint64(21),
				},
			},
		},
		{
			description: "Simple query with basic filter(age), greater than",
			query: `query {
						users(filter: {Age: {_gt: 20}}) {
							Name
							Age
						}
					}`,
			docs: []string{
				(`{
					"Name": "John",
					"Age": 21
				}`),
			},
			results: []map[string]interface{}{
				{
					"Name": "John",
					"Age":  uint64(21),
				},
			},
		},
		{
			description: "Simple query with basic filter(age)",
			query: `query {
						users(filter: {Age: {_gt: 40}}) {
							Name
							Age
						}
					}`,
			docs: []string{
				(`{
					"Name": "John",
					"Age": 21
				}`),
				(`{
					"Name": "Bob",
					"Age": 32
				}`),
			},
			results: []map[string]interface{}{},
		},
		{
			description: "Simple query with basic filter(age)",
			query: `query {
						users(filter: {Age: {_gt: 20}}) {
							Name
							Age
						}
					}`,
			docs: []string{
				(`{
					"Name": "John",
					"Age": 21
				}`),
				(`{
					"Name": "Bob",
					"Age": 32
				}`),
			},
			results: []map[string]interface{}{
				{
					"Name": "Bob",
					"Age":  uint64(32),
				},
				{
					"Name": "John",
					"Age":  uint64(21),
				},
			},
		},
	}

	for _, test := range tests {
		runQueryTestCase(t, col, test)
	}

}

func runQueryTestCase(t *testing.T, collection *Collection, test queryTestCase) {
	// insert docs
	for _, docStr := range test.docs {
		doc, err := document.NewFromJSON([]byte(docStr))
		assert.NoError(t, err, test.description)
		collection.Save(doc)
	}

	// exec query
	txn, err := collection.db.NewTxn(true)
	assert.NoError(t, err, test.description)
	results, err := collection.db.queryExecutor.ExecQuery(txn, test.query)
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

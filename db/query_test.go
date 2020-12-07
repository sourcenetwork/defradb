package db

import (
	"fmt"
	"testing"

	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/query/graphql/planner"

	"github.com/stretchr/testify/assert"
)

var userCollectionGQLSchema = (`
type users {
	Name: String
	Age: Int
}
`)

var userQuery = (`
query {
	users {
		Name
		Age
	}
}`)

// func newQueryableDB()

func TestSimpleCollectionQuery(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)

	desc := newTestCollectionDescription()
	col, err := db.CreateCollection(desc)
	assert.NoError(t, err)

	executor, err := planner.NewQueryExecutor()
	assert.NoError(t, err)

	err = executor.Generator.FromSDL(userCollectionGQLSchema)
	assert.NoError(t, err)

	doc1, err := document.NewFromJSON([]byte(`{
		"Name": "John",
		"Age": 21
	}`))

	assert.NoError(t, err)
	err = col.Save(doc1)
	assert.NoError(t, err)

	txn, err := db.NewTxn(true)
	assert.NoError(t, err)

	// obj := executor.SchemaManager.Schema().TypeMap()["users"].(*gql.Object)
	// obj.Fields()
	// spew.Dump(obj.Fields())

	docs, err := executor.ExecuteQueryInTransaction(txn, userQuery)
	assert.NoError(t, err)

	fmt.Println(docs)
	assert.True(t, len(docs) == 1)
}

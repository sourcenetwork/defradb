package db

import (
	"testing"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/db/fetcher"
	"github.com/sourcenetwork/defradb/document"
	"github.com/stretchr/testify/assert"
)

// func newMemoryDB() (*db.DB, error) {
// 	opts := &db.Options{
// 		Store: "memory",
// 		Memory: db.MemoryOptions{
// 			Size: 1024 * 1000,
// 		},
// 	}

// 	return db.NewDB(opts)
// }

// Create a new Fetcher for a Collection named "users"
// with the following schema:
// Users {
//		Name string
//		Age int
// }

func newTestCollectionDescription() base.CollectionDescription {
	return base.CollectionDescription{
		Name: "users",
		ID:   uint32(1),
		Schema: base.SchemaDescription{
			ID:       uint32(1),
			FieldIDs: []uint32{1, 2, 3},
			Fields: []base.FieldDescription{
				base.FieldDescription{
					Name: "_dockey",
					ID:   uint32(1),
					Kind: base.FieldKind_DocKey,
				},
				base.FieldDescription{
					Name: "Name",
					ID:   uint32(2),
					Kind: base.FieldKind_STRING,
					Typ:  core.LWW_REGISTER,
				},
				base.FieldDescription{
					Name: "Age",
					ID:   uint32(3),
					Kind: base.FieldKind_INT,
					Typ:  core.LWW_REGISTER,
				},
			},
		},
		Indexes: []base.IndexDescription{
			base.IndexDescription{
				Name:    "primary",
				ID:      uint32(0),
				Primary: true,
				Unique:  true,
			},
		},
	}

}

func newTestFetcher() (*fetcher.DocumentFetcher, error) {
	df := new(fetcher.DocumentFetcher)
	desc := newTestCollectionDescription()
	err := df.Init(&desc, &desc.Indexes[0], nil, false)
	if err != nil {
		return nil, err
	}
	return df, nil
}

func TestFetcherInit(t *testing.T) {
	_, err := newTestFetcher()
	assert.NoError(t, err)
}

func TestFetcherStart(t *testing.T) {
	db, err := newMemoryDB()
	if err != nil {
		t.Error(err)
		return
	}
	txn, err := db.NewTxn(true)
	if err != nil {
		t.Error(err)
		return
	}
	df, err := newTestFetcher()
	assert.NoError(t, err)

	err = df.Start(txn, core.Spans{})
	assert.NoError(t, err)
}

func TestFetcherStartWithoutInit(t *testing.T) {
	db, err := newMemoryDB()
	if err != nil {
		t.Error(err)
		return
	}
	txn, err := db.NewTxn(true)
	if err != nil {
		t.Error(err)
		return
	}
	df := new(fetcher.DocumentFetcher)
	err = df.Start(txn, core.Spans{})
	assert.Error(t, err)
}

func TestMakeIndexPrefixKey(t *testing.T) {
	desc := newTestCollectionDescription()
	key := base.MakeIndexPrefixKey(&desc, &desc.Indexes[0])
	assert.Equal(t, "/db/data/1/0", key.String())
}

func TestFetcherGetAllPrimaryIndexEncodedDocSingle(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)

	col, err := newTestCollectionWithSchema(db)
	assert.NoError(t, err)

	doc, err := document.NewFromJSON([]byte(`{
		"Name": "John",
		"Age": 21
	}`))
	assert.NoError(t, err)
	err = col.Save(doc)
	assert.NoError(t, err)

	txn, err := db.NewTxn(true)
	if err != nil {
		t.Error(err)
		return
	}

	// db.printDebugDB()

	df := new(fetcher.DocumentFetcher)
	desc := col.Description()
	err = df.Init(&desc, &desc.Indexes[0], nil, false)
	assert.NoError(t, err)

	err = df.Start(txn, core.Spans{})
	assert.NoError(t, err)

	// assert.False(t, df.KVEnd())
	// kv := df.KV()
	// assert.NotNil(t, kv)
	// fmt.Println(kv)
	// // err = df.ProcessKV(kv)
	// // assert.Nil(t, err)
	// // err = df.NextKey()
	// assert.True(t, false)

	// var _ []*document.EncodedDocument
	encdoc, err := df.FetchNext()
	assert.NoError(t, err)
	assert.NotNil(t, encdoc)

	// fmt.Println(encdoc)
	// assert.True(t, false)
}

func TestFetcherGetAllPrimaryIndexEncodedDocMultiple(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)

	col, err := newTestCollectionWithSchema(db)
	assert.NoError(t, err)

	doc, err := document.NewFromJSON([]byte(`{
		"Name": "John",
		"Age": 21
	}`))
	assert.NoError(t, err)
	err = col.Save(doc)
	assert.NoError(t, err)

	doc, err = document.NewFromJSON([]byte(`{
		"Name": "Alice",
		"Age": 27
	}`))
	assert.NoError(t, err)
	err = col.Save(doc)
	assert.NoError(t, err)

	txn, err := db.NewTxn(true)
	if err != nil {
		t.Error(err)
		return
	}

	// db.printDebugDB()

	df := new(fetcher.DocumentFetcher)
	desc := col.Description()
	err = df.Init(&desc, &desc.Indexes[0], nil, false)
	assert.NoError(t, err)

	err = df.Start(txn, core.Spans{})
	assert.NoError(t, err)

	// assert.False(t, df.KVEnd())
	// kv := df.KV()
	// assert.NotNil(t, kv)
	// fmt.Println(kv)
	// // err = df.ProcessKV(kv)
	// // assert.Nil(t, err)
	// // err = df.NextKey()
	// assert.True(t, false)

	// var _ []*document.EncodedDocument
	encdoc, err := df.FetchNext()
	assert.NoError(t, err)
	assert.NotNil(t, encdoc)
	// fmt.Println(encdoc)
	encdoc, err = df.FetchNext()
	assert.NoError(t, err)
	assert.NotNil(t, encdoc)

	// fmt.Println(encdoc)
	// assert.True(t, false)
}

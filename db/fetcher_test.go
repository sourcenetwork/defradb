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

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/db/fetcher"
	"github.com/sourcenetwork/defradb/document"
	"github.com/stretchr/testify/assert"
)

func newTestCollectionDescription() base.CollectionDescription {
	return base.CollectionDescription{
		Name: "users",
		ID:   uint32(1),
		Schema: base.SchemaDescription{
			ID:       uint32(1),
			FieldIDs: []uint32{1, 2, 3},
			Fields: []base.FieldDescription{
				{
					Name: "_key",
					ID:   base.FieldID(1),
					Kind: base.FieldKind_DocKey,
				},
				{
					Name: "Name",
					ID:   base.FieldID(2),
					Kind: base.FieldKind_STRING,
					Typ:  core.LWW_REGISTER,
				},
				{
					Name: "Age",
					ID:   base.FieldID(3),
					Kind: base.FieldKind_INT,
					Typ:  core.LWW_REGISTER,
				},
			},
		},
		Indexes: []base.IndexDescription{
			{
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

func TestFetcherGetAllPrimaryIndexDecodedSingle(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)

	col, err := newTestCollectionWithSchema(db)
	assert.NoError(t, err)

	txn, err := db.NewTxn(true)
	if err != nil {
		t.Error(err)
		return
	}

	doc, err := document.NewFromJSON([]byte(`{
		"Name": "John",
		"Age": 21
	}`))
	assert.NoError(t, err)
	err = col.Save(doc)
	assert.NoError(t, err)

	df := new(fetcher.DocumentFetcher)
	desc := col.Description()
	err = df.Init(&desc, &desc.Indexes[0], nil, false)
	assert.NoError(t, err)

	err = df.Start(txn, core.Spans{})
	assert.NoError(t, err)

	ddoc, err := df.FetchNextDecoded()
	assert.NoError(t, err)
	assert.NotNil(t, ddoc)

	// value check
	name, err := ddoc.Get("Name")
	assert.NoError(t, err)
	age, err := ddoc.Get("Age")
	assert.NoError(t, err)

	assert.Equal(t, "John", name)
	assert.Equal(t, uint64(21), age)
	fmt.Println(age)
}

func TestFetcherGetAllPrimaryIndexDecodedMultiple(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)

	col, err := newTestCollectionWithSchema(db)
	assert.NoError(t, err)

	txn, err := db.NewTxn(true)
	if err != nil {
		t.Error(err)
		return
	}

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

	df := new(fetcher.DocumentFetcher)
	desc := col.Description()
	err = df.Init(&desc, &desc.Indexes[0], nil, false)
	assert.NoError(t, err)

	err = df.Start(txn, core.Spans{})
	assert.NoError(t, err)

	ddoc, err := df.FetchNextDecoded()
	assert.NoError(t, err)
	assert.NotNil(t, ddoc)

	// value check
	name, err := ddoc.Get("Name")
	assert.NoError(t, err)
	age, err := ddoc.Get("Age")
	assert.NoError(t, err)

	assert.Equal(t, "John", name)
	assert.Equal(t, uint64(21), age)

	ddoc, err = df.FetchNextDecoded()
	assert.NoError(t, err)
	assert.NotNil(t, ddoc)

	// value check
	name, err = ddoc.Get("Name")
	assert.NoError(t, err)
	age, err = ddoc.Get("Age")
	assert.NoError(t, err)

	assert.Equal(t, "Alice", name)
	assert.Equal(t, uint64(27), age)
}

func TestFetcherGetOnePrimaryIndexDecoded(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)

	col, err := newTestCollectionWithSchema(db)
	assert.NoError(t, err)

	txn, err := db.NewTxn(true)
	if err != nil {
		t.Error(err)
		return
	}

	doc, err := document.NewFromJSON([]byte(`{
		"Name": "John",
		"Age": 21
	}`))
	assert.NoError(t, err)
	err = col.Save(doc)
	assert.NoError(t, err)

	df := new(fetcher.DocumentFetcher)
	desc := col.Description()
	err = df.Init(&desc, &desc.Indexes[0], nil, false)
	assert.NoError(t, err)

	// create a span for our document we wish to find
	docKey := core.Key{base.MakeIndexPrefixKey(&desc, &desc.Indexes[0]).ChildString("bae-52b9170d-b77a-5887-b877-cbdbb99b009f")}
	spans := core.Spans{
		core.NewSpan(docKey, docKey.PrefixEnd()),
	}
	err = df.Start(txn, spans)
	assert.NoError(t, err)

	ddoc, err := df.FetchNextDecoded()
	assert.NoError(t, err)
	assert.NotNil(t, ddoc)

	// value check
	name, err := ddoc.Get("Name")
	assert.NoError(t, err)
	age, err := ddoc.Get("Age")
	assert.NoError(t, err)

	assert.Equal(t, "John", name)
	assert.Equal(t, uint64(21), age)
	// fmt.Println(age)

	// db.printDebugDB()
	// assert.True(t, false)
}

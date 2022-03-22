// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package fetcher_test

import (
	"context"
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/db/fetcher"
	"github.com/stretchr/testify/assert"
)

func newTestCollectionDescription() client.CollectionDescription {
	return client.CollectionDescription{
		Name: "users",
		ID:   uint32(1),
		Schema: client.SchemaDescription{
			Fields: []client.FieldDescription{
				{
					Name: "_key",
					ID:   client.FieldID(1),
					Kind: client.FieldKind_DocKey,
				},
				{
					Name: "name",
					ID:   client.FieldID(2),
					Kind: client.FieldKind_STRING,
					Typ:  client.LWW_REGISTER,
				},
				{
					Name: "age",
					ID:   client.FieldID(3),
					Kind: client.FieldKind_INT,
					Typ:  client.LWW_REGISTER,
				},
			},
		},
		Indexes: []client.IndexDescription{
			{
				ID: uint32(0),
			},
		},
	}

}

func newTestFetcher() (*fetcher.DocumentFetcher, error) {
	df := new(fetcher.DocumentFetcher)
	desc := newTestCollectionDescription()
	err := df.Init(&desc, &desc.Indexes[0])
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
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	txn, err := db.NewTxn(ctx, true)
	if err != nil {
		t.Error(err)
		return
	}
	df, err := newTestFetcher()
	assert.NoError(t, err)

	err = df.Start(ctx, txn, core.Spans{})
	assert.NoError(t, err)
}

func TestFetcherStartWithoutInit(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	txn, err := db.NewTxn(ctx, true)
	if err != nil {
		t.Error(err)
		return
	}
	df := new(fetcher.DocumentFetcher)
	err = df.Start(ctx, txn, core.Spans{})
	assert.Error(t, err)
}

func TestMakeIndexPrefixKey(t *testing.T) {
	desc := newTestCollectionDescription()
	key := base.MakeIndexPrefixKey(desc, &desc.Indexes[0])
	assert.Equal(t, "/1/0", key.ToString())
}

func TestFetcherGetAllPrimaryIndexEncodedDocSingle(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	col, err := newTestCollectionWithSchema(db)
	assert.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{
		"name": "John",
		"age": 21
	}`))
	assert.NoError(t, err)
	err = col.Save(ctx, doc)
	assert.NoError(t, err)

	txn, err := db.NewTxn(ctx, true)
	if err != nil {
		t.Error(err)
		return
	}

	// db.printDebugDB()

	df := new(fetcher.DocumentFetcher)
	desc := col.Description()
	err = df.Init(&desc, &desc.Indexes[0])
	assert.NoError(t, err)

	err = df.Start(ctx, txn, core.Spans{})
	assert.NoError(t, err)

	encdoc, err := df.FetchNext(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, encdoc)
}

func TestFetcherGetAllPrimaryIndexEncodedDocMultiple(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	col, err := newTestCollectionWithSchema(db)
	assert.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{
		"name": "John",
		"age": 21
	}`))
	assert.NoError(t, err)
	err = col.Save(ctx, doc)
	assert.NoError(t, err)

	doc, err = client.NewDocFromJSON([]byte(`{
		"name": "Alice",
		"age": 27
	}`))
	assert.NoError(t, err)
	err = col.Save(ctx, doc)
	assert.NoError(t, err)

	txn, err := db.NewTxn(ctx, true)
	if err != nil {
		t.Error(err)
		return
	}

	// db.printDebugDB()

	df := new(fetcher.DocumentFetcher)
	desc := col.Description()
	err = df.Init(&desc, &desc.Indexes[0])
	assert.NoError(t, err)

	err = df.Start(ctx, txn, core.Spans{})
	assert.NoError(t, err)

	encdoc, err := df.FetchNext(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, encdoc)
	encdoc, err = df.FetchNext(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, encdoc)
}

func TestFetcherGetAllPrimaryIndexDecodedSingle(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	col, err := newTestCollectionWithSchema(db)
	assert.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{
		"name": "John",
		"age": 21
	}`))
	assert.NoError(t, err)
	err = col.Save(ctx, doc)
	assert.NoError(t, err)

	df := new(fetcher.DocumentFetcher)
	desc := col.Description()
	err = df.Init(&desc, &desc.Indexes[0])
	assert.NoError(t, err)

	txn, err := db.NewTxn(ctx, true)
	if err != nil {
		t.Error(err)
		return
	}

	err = df.Start(ctx, txn, core.Spans{})
	assert.NoError(t, err)

	ddoc, err := df.FetchNextDecoded(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, ddoc)

	// value check
	name, err := ddoc.Get("name")
	assert.NoError(t, err)
	age, err := ddoc.Get("age")
	assert.NoError(t, err)

	assert.Equal(t, "John", name)
	assert.Equal(t, uint64(21), age)
}

func TestFetcherGetAllPrimaryIndexDecodedMultiple(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	col, err := newTestCollectionWithSchema(db)
	assert.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{
		"name": "John",
		"age": 21
	}`))
	assert.NoError(t, err)
	err = col.Save(ctx, doc)
	assert.NoError(t, err)

	doc, err = client.NewDocFromJSON([]byte(`{
		"name": "Alice",
		"age": 27
	}`))
	assert.NoError(t, err)
	err = col.Save(ctx, doc)
	assert.NoError(t, err)

	df := new(fetcher.DocumentFetcher)
	desc := col.Description()
	err = df.Init(&desc, &desc.Indexes[0])
	assert.NoError(t, err)

	txn, err := db.NewTxn(ctx, true)
	if err != nil {
		t.Error(err)
		return
	}

	err = df.Start(ctx, txn, core.Spans{})
	assert.NoError(t, err)

	ddoc, err := df.FetchNextDecoded(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, ddoc)

	// value check
	name, err := ddoc.Get("name")
	assert.NoError(t, err)
	age, err := ddoc.Get("age")
	assert.NoError(t, err)

	assert.Equal(t, "Alice", name)
	assert.Equal(t, uint64(27), age)

	ddoc, err = df.FetchNextDecoded(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, ddoc)

	// value check
	name, err = ddoc.Get("name")
	assert.NoError(t, err)
	age, err = ddoc.Get("age")
	assert.NoError(t, err)

	assert.Equal(t, "John", name)
	assert.Equal(t, uint64(21), age)
}

func TestFetcherGetOnePrimaryIndexDecoded(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	col, err := newTestCollectionWithSchema(db)
	assert.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{
		"name": "John",
		"age": 21
	}`))
	assert.NoError(t, err)
	err = col.Save(ctx, doc)
	assert.NoError(t, err)

	df := new(fetcher.DocumentFetcher)
	desc := col.Description()
	err = df.Init(&desc, &desc.Indexes[0])
	assert.NoError(t, err)

	// create a span for our document we wish to find
	docKey := base.MakeIndexPrefixKey(desc, &desc.Indexes[0]).WithDocKey("bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7")
	spans := core.Spans{
		core.NewSpan(docKey, docKey.PrefixEnd()),
	}

	txn, err := db.NewTxn(ctx, true)
	if err != nil {
		t.Error(err)
		return
	}

	err = df.Start(ctx, txn, spans)
	assert.NoError(t, err)

	ddoc, err := df.FetchNextDecoded(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, ddoc)

	// value check
	name, err := ddoc.Get("name")
	assert.NoError(t, err)
	age, err := ddoc.Get("age")
	assert.NoError(t, err)

	assert.Equal(t, "John", name)
	assert.Equal(t, uint64(21), age)
}

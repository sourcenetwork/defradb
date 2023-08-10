// Copyright 2022 Democratized Data Foundation
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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/db/fetcher"
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
					Name: "Name",
					ID:   client.FieldID(2),
					Kind: client.FieldKind_STRING,
					Typ:  client.LWW_REGISTER,
				},
				{
					Name: "Age",
					ID:   client.FieldID(3),
					Kind: client.FieldKind_INT,
					Typ:  client.LWW_REGISTER,
				},
			},
		},
	}
}

func newTestFetcher(ctx context.Context, txn datastore.Txn) (*fetcher.DocumentFetcher, error) {
	df := new(fetcher.DocumentFetcher)
	desc := newTestCollectionDescription()
	err := df.Init(ctx, txn, &desc, desc.Schema.Fields, nil, nil, false, false)
	if err != nil {
		return nil, err
	}
	return df, nil
}

func TestFetcherInit(t *testing.T) {
	_, err := newTestFetcher(context.Background(), nil)
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
	df, err := newTestFetcher(ctx, txn)
	assert.NoError(t, err)

	err = df.Start(ctx, core.Spans{})
	assert.NoError(t, err)
}

func TestFetcherStartWithoutInit(t *testing.T) {
	ctx := context.Background()
	df := new(fetcher.DocumentFetcher)
	err := df.Start(ctx, core.Spans{})
	assert.Error(t, err)
}

func TestMakeIndexPrefixKey(t *testing.T) {
	desc := newTestCollectionDescription()
	key := base.MakeCollectionKey(desc)
	assert.Equal(t, "/1", key.ToString())
}

func TestFetcherGetAllPrimaryIndexEncodedDocSingle(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	col, err := newTestCollectionWithSchema(t, ctx, db)
	assert.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{
		"Name": "John",
		"Age": 21
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
	err = df.Init(ctx, txn, &desc, desc.Schema.Fields, nil, nil, false, false)
	assert.NoError(t, err)

	err = df.Start(ctx, core.Spans{})
	assert.NoError(t, err)

	encdoc, _, err := df.FetchNext(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, encdoc)
}

func TestFetcherGetAllPrimaryIndexEncodedDocMultiple(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	col, err := newTestCollectionWithSchema(t, ctx, db)
	assert.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{
		"Name": "John",
		"Age": 21
	}`))
	assert.NoError(t, err)
	err = col.Save(ctx, doc)
	assert.NoError(t, err)

	doc, err = client.NewDocFromJSON([]byte(`{
		"Name": "Alice",
		"Age": 27
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
	err = df.Init(ctx, txn, &desc, desc.Schema.Fields, nil, nil, false, false)
	assert.NoError(t, err)

	err = df.Start(ctx, core.Spans{})
	assert.NoError(t, err)

	encdoc, _, err := df.FetchNext(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, encdoc)
	encdoc, _, err = df.FetchNext(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, encdoc)
}

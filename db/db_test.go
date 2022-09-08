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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"

	corecrdt "github.com/sourcenetwork/defradb/core/crdt"
	"github.com/sourcenetwork/defradb/merkle/clock"

	badger "github.com/dgraph-io/badger/v3"
	ds "github.com/ipfs/go-datastore"
	dag "github.com/ipfs/go-merkledag"
	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v3"
	"github.com/stretchr/testify/assert"
)

func newMemoryDB(ctx context.Context) (*db, error) {
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	if err != nil {
		return nil, err
	}
	return newDB(ctx, rootstore)
}

func TestNewDB(t *testing.T) {
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = NewDB(ctx, rootstore)
	if err != nil {
		t.Error(err)
	}
}

func TestNewDBWithCollection_Errors_GivenNoSchema(t *testing.T) {
	ctx := context.Background()
	rootstore := ds.NewMapDatastore()

	db, err := NewDB(ctx, rootstore)
	if err != nil {
		t.Error(err)
	}

	_, err = db.CreateCollection(ctx, client.CollectionDescription{
		Name: "test",
	})

	assert.Error(t, err)
}

func TestDBSaveSimpleDocument(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)
	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21,
		"Weight": 154.1
	}`)

	doc, err := client.NewDocFromJSON(testJSONObj)
	if err != nil {
		t.Error(err)
		return
	}

	err = col.Save(ctx, doc)
	if err != nil {
		t.Error(err)
	}

	// value check
	name, err := doc.Get("Name")
	assert.NoError(t, err)
	age, err := doc.Get("Age")
	assert.NoError(t, err)
	weight, err := doc.Get("Weight")
	assert.NoError(t, err)

	assert.Equal(t, "John", name)
	assert.Equal(t, int64(21), age)
	assert.Equal(t, 154.1, weight)

	_, err = doc.Get("DoesntExist")
	assert.Error(t, err)

	// db.printDebugDB()
}

func TestDBUpdateDocument(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)
	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21,
		"Weight": 154.1
	}`)

	doc, err := client.NewDocFromJSON(testJSONObj)
	if err != nil {
		t.Error(err)
		return
	}

	err = col.Save(ctx, doc)
	if err != nil {
		t.Error(err)
	}

	// update fields
	doc.Set("Name", "Pete")
	doc.Delete("Weight")

	weightField := doc.Fields()["Weight"]
	weightVal, _ := doc.GetValueWithField(weightField)
	assert.True(t, weightVal.IsDelete())

	err = col.Update(ctx, doc)
	if err != nil {
		t.Error(err)
	}

	// value check
	name, err := doc.Get("Name")
	assert.NoError(t, err)
	age, err := doc.Get("Age")
	assert.NoError(t, err)
	weight, err := doc.Get("Weight")
	assert.NoError(t, err)

	assert.Equal(t, "Pete", name)
	assert.Equal(t, int64(21), age)
	assert.Nil(t, weight)
}

func TestDBUpdateNonExistingDocument(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)
	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21,
		"Weight": 154.1
	}`)

	doc, err := client.NewDocFromJSON(testJSONObj)
	if err != nil {
		t.Error(err)
		return
	}

	err = col.Update(ctx, doc)
	assert.Error(t, err)
}

func TestDBUpdateExistingDocument(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)
	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21,
		"Weight": 154.1
	}`)

	doc, err := client.NewDocFromJSON(testJSONObj)
	assert.NoError(t, err)

	err = col.Save(ctx, doc)
	assert.NoError(t, err)

	testJSONObj = []byte(`{
		"_key": "bae-09cd7539-9b86-5661-90f6-14fbf6c1a14d",
		"Name": "Pete",
		"Age": 31
	}`)

	doc, err = client.NewDocFromJSON(testJSONObj)
	assert.NoError(t, err)

	err = col.Update(ctx, doc)
	assert.NoError(t, err)

	// value check
	name, err := doc.Get("Name")
	assert.NoError(t, err)
	age, err := doc.Get("Age")
	assert.NoError(t, err)
	// weight, err := doc.Get("Weight")
	// assert.NoError(t, err)

	assert.Equal(t, "Pete", name)
	assert.Equal(t, int64(31), age)
}

func TestDBGetDocument(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)
	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21,
		"Weight": 154.1
	}`)

	doc, err := client.NewDocFromJSON(testJSONObj)
	assert.NoError(t, err)

	err = col.Save(ctx, doc)
	assert.NoError(t, err)

	key, err := client.NewDocKeyFromString("bae-09cd7539-9b86-5661-90f6-14fbf6c1a14d")
	assert.NoError(t, err)
	doc, err = col.Get(ctx, key)
	assert.NoError(t, err)

	// value check
	name, err := doc.Get("Name")
	assert.NoError(t, err)
	age, err := doc.Get("Age")
	assert.NoError(t, err)
	weight, err := doc.Get("Weight")
	assert.NoError(t, err)

	assert.Equal(t, "John", name)
	assert.Equal(
		t,
		uint64(21),
		age,
	) // note: uint is used here, because the CBOR implementation converts all positive ints to uint64
	assert.Equal(t, 154.1, weight)
}

func TestDBGetNotFoundDocument(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)
	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	key, err := client.NewDocKeyFromString("bae-09cd7539-9b86-5661-90f6-14fbf6c1a14d")
	assert.NoError(t, err)
	_, err = col.Get(ctx, key)
	assert.EqualError(t, err, client.ErrDocumentNotFound.Error())
}

func TestDBDeleteDocument(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)
	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21,
		"Weight": 154.1
	}`)

	doc, err := client.NewDocFromJSON(testJSONObj)
	assert.NoError(t, err)

	err = col.Save(ctx, doc)
	assert.NoError(t, err)

	key, err := client.NewDocKeyFromString("bae-09cd7539-9b86-5661-90f6-14fbf6c1a14d")
	assert.NoError(t, err)
	deleted, err := col.Delete(ctx, key)
	assert.NoError(t, err)
	assert.True(t, deleted)
}

func TestDBDeleteNotFoundDocument(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)
	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	key, err := client.NewDocKeyFromString("bae-09cd7539-9b86-5661-90f6-14fbf6c1a14d")
	assert.NoError(t, err)
	deleted, err := col.Delete(ctx, key)
	assert.EqualError(t, err, client.ErrDocumentNotFound.Error())
	assert.False(t, deleted)
}

func TestDocumentMerkleDAG(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)
	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21,
		"Weight": 154.1
	}`)

	doc, err := client.NewDocFromJSON(testJSONObj)
	assert.NoError(t, err)

	err = col.Save(ctx, doc)
	assert.NoError(t, err)

	clk := clock.NewMerkleClock(
		db.multistore.Headstore(),
		nil,
		core.HeadStoreKey{}.WithDocKey(
			"bae-09cd7539-9b86-5661-90f6-14fbf6c1a14d",
		).WithFieldId(
			"Name",
		),
		nil,
	)
	heads := clk.(*clock.MerkleClock).Heads()
	cids, _, err := heads.List(ctx)
	assert.NoError(t, err)

	reg := corecrdt.LWWRegister{}
	for _, c := range cids {
		b, errGet := db.Blockstore().Get(ctx, c)
		assert.NoError(t, errGet)

		nd, errDecode := dag.DecodeProtobuf(b.RawData())
		assert.NoError(t, errDecode)

		_, errMarshal := nd.MarshalJSON()
		assert.NoError(t, errMarshal)

		_, errDeltaDecode := reg.DeltaDecode(nd)
		assert.NoError(t, errDeltaDecode)
	}

	testJSONObj = []byte(`{
		"_key": "bae-09cd7539-9b86-5661-90f6-14fbf6c1a14d",
		"Name": "Pete",
		"Age": 31
	}`)

	doc, err = client.NewDocFromJSON(testJSONObj)
	assert.NoError(t, err)

	err = col.Update(ctx, doc)
	assert.NoError(t, err)

	heads = clk.(*clock.MerkleClock).Heads()
	cids, _, err = heads.List(ctx)
	assert.NoError(t, err)

	for _, c := range cids {
		b, err := db.Blockstore().Get(ctx, c)
		assert.NoError(t, err)

		nd, err := dag.DecodeProtobuf(b.RawData())
		assert.NoError(t, err)

		_, err = nd.MarshalJSON()
		assert.NoError(t, err)

		_, err = reg.DeltaDecode(nd)
		assert.NoError(t, err)
	}
}

// collection with schema
func TestDBSchemaSaveSimpleDocument(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)
	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21
	}`)

	doc, err := client.NewDocFromJSON(testJSONObj)
	if err != nil {
		t.Error(err)
		return
	}

	err = col.Save(ctx, doc)
	assert.NoError(t, err)

	// value check
	name, err := doc.Get("Name")
	assert.NoError(t, err)
	age, err := doc.Get("Age")
	assert.NoError(t, err)

	assert.Equal(t, "John", name)
	assert.Equal(t, int64(21), age)

	err = db.PrintDump(ctx)
	assert.Nil(t, err)
}

func TestDBUpdateDocWithKeyAndInvalidJSONString(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)
	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21
	}`)

	doc, err := client.NewDocFromJSON(testJSONObj)
	if err != nil {
		t.Error(err)
		return
	}

	err = col.Save(ctx, doc)
	assert.NoError(t, err)

	_, err = col.UpdateWithKey(ctx, doc.Key(), `{{
		"Name": "Eric"
	}`)
	assert.ErrorContains(t, err, "cannot parse JSON")

	doc, err = col.Get(ctx, doc.Key())
	assert.NoError(t, err)

	name, err := doc.Get("Name")
	assert.NoError(t, err)
	assert.Equal(t, "John", name)
}

func TestDBUpdateDocWithKeyAndInvalidUpdator(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)
	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21
	}`)

	doc, err := client.NewDocFromJSON(testJSONObj)
	if err != nil {
		t.Error(err)
		return
	}

	err = col.Save(ctx, doc)
	assert.NoError(t, err)

	_, err = col.UpdateWithKey(ctx, doc.Key(), `"Name: Eric"`)
	assert.ErrorIs(t, err, client.ErrInvalidUpdater)

	doc, err = col.Get(ctx, doc.Key())
	assert.NoError(t, err)

	name, err := doc.Get("Name")
	assert.NoError(t, err)
	assert.Equal(t, "John", name)
}

func TestDBUpdateDocWithKeyAndPatchUpdator(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)
	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21
	}`)

	doc, err := client.NewDocFromJSON(testJSONObj)
	if err != nil {
		t.Error(err)
		return
	}

	err = col.Save(ctx, doc)
	assert.NoError(t, err)

	_, err = col.UpdateWithKey(ctx, doc.Key(), `[
		{
			"Name": "Eric"
		}, {
			"Name": "Sam"
		}
	]`)
	assert.NoError(t, err)

	doc, err = col.Get(ctx, doc.Key())
	assert.NoError(t, err)

	name, err := doc.Get("Name")
	assert.NoError(t, err)
	// The value of name will not have change since the patch code path is not imlpemented.
	assert.Equal(t, "John", name)
}

func TestDBUpdateDocWithKey(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)
	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21
	}`)

	doc, err := client.NewDocFromJSON(testJSONObj)
	if err != nil {
		t.Error(err)
		return
	}

	err = col.Save(ctx, doc)
	assert.NoError(t, err)

	_, err = col.UpdateWithKey(ctx, doc.Key(), `{
		"Name": "Eric"
	}`)
	assert.NoError(t, err)

	doc, err = col.Get(ctx, doc.Key())
	assert.NoError(t, err)

	name, err := doc.Get("Name")
	assert.NoError(t, err)
	assert.Equal(t, "Eric", name)
}

func TestDBUpdateDocWithKeysAndInvalidJSONString(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)
	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21
	}`)

	doc, err := client.NewDocFromJSON(testJSONObj)
	if err != nil {
		t.Error(err)
		return
	}

	err = col.Save(ctx, doc)
	assert.NoError(t, err)

	_, err = col.UpdateWithKey(ctx, doc.Key(), `{{
		"Name": "Eric"
	}`)
	assert.ErrorContains(t, err, "cannot parse JSON")

	doc, err = col.Get(ctx, doc.Key())
	assert.NoError(t, err)

	name, err := doc.Get("Name")
	assert.NoError(t, err)
	assert.Equal(t, "John", name)
}

func TestDBUpdateDocWithKeysAndInvalidUpdator(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)
	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	testJSONObj1 := []byte(`{
		"Name": "John",
		"Age": 21
	}`)

	doc1, err := client.NewDocFromJSON(testJSONObj1)
	if err != nil {
		t.Error(err)
		return
	}

	err = col.Save(ctx, doc1)
	assert.NoError(t, err)

	testJSONObj2 := []byte(`{
		"Name": "Sam",
		"Age": 32
	}`)

	doc2, err := client.NewDocFromJSON(testJSONObj2)
	if err != nil {
		t.Error(err)
		return
	}

	err = col.Save(ctx, doc2)
	assert.NoError(t, err)

	_, err = col.UpdateWithKeys(ctx, []client.DocKey{doc1.Key(), doc2.Key()}, `"Age: 40"`)
	assert.ErrorIs(t, err, client.ErrInvalidUpdater)

	doc1, err = col.Get(ctx, doc1.Key())
	assert.NoError(t, err)

	age1, err := doc1.Get("Age")
	assert.NoError(t, err)
	assert.Equal(t, uint64(21), age1)

	doc2, err = col.Get(ctx, doc2.Key())
	assert.NoError(t, err)

	age2, err := doc2.Get("Age")
	assert.NoError(t, err)
	assert.Equal(t, uint64(32), age2)
}

func TestDBUpdateDocWithKeysAndPatchUpdator(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)
	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	testJSONObj1 := []byte(`{
		"Name": "John",
		"Age": 21
	}`)

	doc1, err := client.NewDocFromJSON(testJSONObj1)
	if err != nil {
		t.Error(err)
		return
	}

	err = col.Save(ctx, doc1)
	assert.NoError(t, err)

	testJSONObj2 := []byte(`{
		"Name": "Sam",
		"Age": 32
	}`)

	doc2, err := client.NewDocFromJSON(testJSONObj2)
	if err != nil {
		t.Error(err)
		return
	}

	err = col.Save(ctx, doc2)
	assert.NoError(t, err)

	_, err = col.UpdateWithKeys(ctx, []client.DocKey{doc1.Key(), doc2.Key()}, `[
		{
			"Age": 33
		}, {
			"Age": 40
		}
	]`)
	assert.NoError(t, err)

	doc1, err = col.Get(ctx, doc1.Key())
	assert.NoError(t, err)

	age1, err := doc1.Get("Age")
	assert.NoError(t, err)
	// The value of age will not have change since the patch code path is not imlpemented.
	assert.Equal(t, uint64(21), age1)

	doc2, err = col.Get(ctx, doc2.Key())
	assert.NoError(t, err)

	age2, err := doc2.Get("Age")
	assert.NoError(t, err)
	// The value of age will not have change since the patch code path is not imlpemented.
	assert.Equal(t, uint64(32), age2)
}

func TestDBUpdateDocWithKeys(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)
	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	testJSONObj1 := []byte(`{
		"Name": "John",
		"Age": 21
	}`)

	doc1, err := client.NewDocFromJSON(testJSONObj1)
	if err != nil {
		t.Error(err)
		return
	}

	err = col.Save(ctx, doc1)
	assert.NoError(t, err)

	testJSONObj2 := []byte(`{
		"Name": "Sam",
		"Age": 32
	}`)

	doc2, err := client.NewDocFromJSON(testJSONObj2)
	if err != nil {
		t.Error(err)
		return
	}

	err = col.Save(ctx, doc2)
	assert.NoError(t, err)

	_, err = col.UpdateWithKeys(ctx, []client.DocKey{doc1.Key(), doc2.Key()}, `{
		"Age": 40
	}`)
	assert.NoError(t, err)

	doc1, err = col.Get(ctx, doc1.Key())
	assert.NoError(t, err)

	age1, err := doc1.Get("Age")
	assert.NoError(t, err)
	assert.Equal(t, uint64(40), age1)

	doc2, err = col.Get(ctx, doc2.Key())
	assert.NoError(t, err)

	age2, err := doc2.Get("Age")
	assert.NoError(t, err)
	assert.Equal(t, uint64(40), age2)
}

func TestDBUpdateDocWithFilterAndInvalidJSONString(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)
	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21
	}`)

	doc, err := client.NewDocFromJSON(testJSONObj)
	if err != nil {
		t.Error(err)
		return
	}

	err = col.Save(ctx, doc)
	assert.NoError(t, err)

	_, err = col.UpdateWithFilter(ctx, `{Name: {_eq: "John"}}`, `{{
		"Name": "Eric"
	}`)
	assert.ErrorContains(t, err, "cannot parse JSON")

	doc, err = col.Get(ctx, doc.Key())
	assert.NoError(t, err)

	name, err := doc.Get("Name")
	assert.NoError(t, err)
	assert.Equal(t, "John", name)
}

func TestDBUpdateDocWithFilterAndInvalidUpdator(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)
	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21
	}`)

	doc, err := client.NewDocFromJSON(testJSONObj)
	if err != nil {
		t.Error(err)
		return
	}

	err = col.Save(ctx, doc)
	assert.NoError(t, err)

	_, err = col.UpdateWithFilter(ctx, `{Name: {_eq: "John"}}`, `"Name: Eric"`)
	assert.ErrorIs(t, err, client.ErrInvalidUpdater)

	doc, err = col.Get(ctx, doc.Key())
	assert.NoError(t, err)

	name, err := doc.Get("Name")
	assert.NoError(t, err)
	assert.Equal(t, "John", name)
}

func TestDBUpdateDocWithFilterAndPatchUpdator(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)
	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21
	}`)

	doc, err := client.NewDocFromJSON(testJSONObj)
	if err != nil {
		t.Error(err)
		return
	}

	err = col.Save(ctx, doc)
	assert.NoError(t, err)

	_, err = col.UpdateWithFilter(ctx, `{Name: {_eq: "John"}}`, `[
		{
			"Name": "Eric"
		}, {
			"Name": "Sam"
		}
	]`)
	assert.NoError(t, err)

	doc, err = col.Get(ctx, doc.Key())
	assert.NoError(t, err)

	name, err := doc.Get("Name")
	assert.NoError(t, err)
	assert.Equal(t, "John", name)
}

func TestDBUpdateDocWithFilter(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)
	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21
	}`)

	doc, err := client.NewDocFromJSON(testJSONObj)
	if err != nil {
		t.Error(err)
		return
	}

	err = col.Save(ctx, doc)
	assert.NoError(t, err)

	_, err = col.UpdateWithFilter(ctx, `{Name: {_eq: "John"}}`, `{
		"Name": "Eric"
	}`)
	assert.NoError(t, err)

	doc, err = col.Get(ctx, doc.Key())
	assert.NoError(t, err)

	name, err := doc.Get("Name")
	assert.NoError(t, err)
	assert.Equal(t, "Eric", name)
}

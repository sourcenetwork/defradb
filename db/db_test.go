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

	corecrdt "github.com/sourcenetwork/defradb/core/crdt"
	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/document/key"
	"github.com/sourcenetwork/defradb/merkle/clock"

	dag "github.com/ipfs/go-merkledag"
	"github.com/stretchr/testify/assert"
)

func newMemoryDB() (*DB, error) {
	opts := &Options{
		Store: "memory",
		Memory: MemoryOptions{
			Size: 1024 * 1000,
		},
		Badger: BadgerOptions{
			Path: "test",
		},
	}

	return NewDB(opts)
}

func newTestCollection(db *DB) (*Collection, error) {
	col, err := db.CreateCollection(base.CollectionDescription{
		Name: "test",
	})
	return col.(*Collection), err
}

func TestNewDB(t *testing.T) {
	opts := &Options{
		Store: "memory",
		Memory: MemoryOptions{
			Size: 1024 * 1000,
		},
	}

	_, err := NewDB(opts)
	if err != nil {
		t.Error(err)
	}
}

func TestNewDBWithCollection(t *testing.T) {
	opts := &Options{
		Store: "memory",
		Memory: MemoryOptions{
			Size: 1024 * 1000,
		},
	}

	db, err := NewDB(opts)
	if err != nil {
		t.Error(err)
	}

	_, err = db.CreateCollection(base.CollectionDescription{
		Name: "test",
	})

	if err != nil {
		t.Error(err)
	}
}

func TestDBSaveSimpleDocument(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)
	col, err := newTestCollection(db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21,
		"Weight": 154.1
	}`)

	doc, err := document.NewFromJSON(testJSONObj)
	if err != nil {
		t.Error(err)
		return
	}

	err = col.Save(doc)
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
	db, err := newMemoryDB()
	assert.NoError(t, err)
	col, err := newTestCollection(db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21,
		"Weight": 154.1
	}`)

	doc, err := document.NewFromJSON(testJSONObj)
	if err != nil {
		t.Error(err)
		return
	}

	err = col.Save(doc)
	if err != nil {
		t.Error(err)
	}

	// update fields
	doc.Set("Name", "Pete")
	doc.Delete("Weight")

	weightField := doc.Fields()["Weight"]
	weightVal, _ := doc.GetValueWithField(weightField)
	assert.True(t, weightVal.IsDelete())

	err = col.Update(doc)

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

	// fmt.Println("\n--")
	// db.printDebugDB()
}

func TestDBUpdateNonExistingDocument(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)
	col, err := newTestCollection(db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21,
		"Weight": 154.1
	}`)

	doc, err := document.NewFromJSON(testJSONObj)
	if err != nil {
		t.Error(err)
		return
	}

	err = col.Update(doc)
	assert.Error(t, err)
}

func TestDBUpdateExistingDocument(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)
	col, err := newTestCollection(db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21,
		"Weight": 154.1
	}`)

	doc, err := document.NewFromJSON(testJSONObj)
	assert.NoError(t, err)

	err = col.Save(doc)
	assert.NoError(t, err)

	testJSONObj = []byte(`{
		"_key": "bae-09cd7539-9b86-5661-90f6-14fbf6c1a14d",
		"Name": "Pete",
		"Age": 31
	}`)

	doc, err = document.NewFromJSON(testJSONObj)
	assert.NoError(t, err)

	err = col.Update(doc)
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
	db, err := newMemoryDB()
	assert.NoError(t, err)
	col, err := newTestCollection(db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21,
		"Weight": 154.1
	}`)

	doc, err := document.NewFromJSON(testJSONObj)
	assert.NoError(t, err)

	err = col.Save(doc)
	fmt.Println(doc.Get("Name"))
	assert.NoError(t, err)

	fmt.Printf("-------\n")
	db.printDebugDB()
	fmt.Printf("-------\n")

	key, err := key.NewFromString("bae-09cd7539-9b86-5661-90f6-14fbf6c1a14d")
	assert.NoError(t, err)
	doc, err = col.Get(key)
	fmt.Println(doc)
	assert.NoError(t, err)

	// value check
	name, err := doc.Get("Name")
	fmt.Println("-----------------------------------------------")
	fmt.Println(name)
	assert.NoError(t, err)
	age, err := doc.Get("Age")
	assert.NoError(t, err)
	weight, err := doc.Get("Weight")
	assert.NoError(t, err)

	assert.Equal(t, "John", name)
	assert.Equal(t, uint64(21), age) // note: uint is used here, because the CBOR implementation converts all positive ints to uint64
	assert.Equal(t, 154.1, weight)
}

func TestDBGetNotFoundDocument(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)
	col, err := newTestCollection(db)
	assert.NoError(t, err)

	key, err := key.NewFromString("bae-09cd7539-9b86-5661-90f6-14fbf6c1a14d")
	assert.NoError(t, err)
	_, err = col.Get(key)
	assert.EqualError(t, err, ErrDocumentNotFound.Error())
}

func TestDBDeleteDocument(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)
	col, err := newTestCollection(db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21,
		"Weight": 154.1
	}`)

	doc, err := document.NewFromJSON(testJSONObj)
	assert.NoError(t, err)

	err = col.Save(doc)
	assert.NoError(t, err)

	key, err := key.NewFromString("bae-09cd7539-9b86-5661-90f6-14fbf6c1a14d")
	assert.NoError(t, err)
	deleted, err := col.Delete(key)
	assert.NoError(t, err)
	assert.True(t, deleted)
}

func TestDBDeleteNotFoundDocument(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)
	col, err := newTestCollection(db)
	assert.NoError(t, err)

	key, err := key.NewFromString("bae-09cd7539-9b86-5661-90f6-14fbf6c1a14d")
	assert.NoError(t, err)
	deleted, err := col.Delete(key)
	assert.EqualError(t, err, ErrDocumentNotFound.Error())
	assert.False(t, deleted)
}

func TestDocumentMerkleDAG(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)
	col, err := newTestCollection(db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21,
		"Weight": 154.1
	}`)

	doc, err := document.NewFromJSON(testJSONObj)
	assert.NoError(t, err)

	err = col.Save(doc)
	assert.NoError(t, err)

	clk := clock.NewMerkleClock(db.headstore, nil, core.HeadStoreKey{}.WithDocKey("bae-09cd7539-9b86-5661-90f6-14fbf6c1a14d").WithFieldId("Name"), nil)
	heads := clk.(*clock.MerkleClock).Heads()
	cids, _, err := heads.List()
	assert.NoError(t, err)

	fmt.Printf("-------\n")
	db.printDebugDB()
	fmt.Printf("-------\n")

	reg := corecrdt.LWWRegister{}
	for _, c := range cids {
		b, err := db.dagstore.Get(c)
		assert.NoError(t, err)

		nd, err := dag.DecodeProtobuf(b.RawData())
		assert.NoError(t, err)
		buf, err := nd.MarshalJSON()
		assert.NoError(t, err)
		fmt.Println(string(buf))
		delta, err := reg.DeltaDecode(nd)
		lwwdelta := delta.(*corecrdt.LWWRegDelta)
		fmt.Printf("%+v - %v\n", lwwdelta, string(lwwdelta.Data))
	}

	testJSONObj = []byte(`{
		"_key": "bae-09cd7539-9b86-5661-90f6-14fbf6c1a14d",
		"Name": "Pete",
		"Age": 31
	}`)

	doc, err = document.NewFromJSON(testJSONObj)
	assert.NoError(t, err)

	err = col.Update(doc)
	assert.NoError(t, err)

	heads = clk.(*clock.MerkleClock).Heads()
	cids, _, err = heads.List()
	assert.NoError(t, err)

	fmt.Printf("-------\n")
	db.printDebugDB()
	fmt.Printf("-------\n")

	for _, c := range cids {
		b, err := db.dagstore.Get(c)
		assert.NoError(t, err)

		nd, err := dag.DecodeProtobuf(b.RawData())
		assert.NoError(t, err)
		buf, err := nd.MarshalJSON()
		assert.NoError(t, err)
		fmt.Println(string(buf))
		delta, err := reg.DeltaDecode(nd)
		lwwdelta := delta.(*corecrdt.LWWRegDelta)
		fmt.Printf("%+v - %v\n", lwwdelta, string(lwwdelta.Data))
	}
}

// collection with schema
func TestDBSchemaSaveSimpleDocument(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)
	col, err := newTestCollectionWithSchema(db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21
	}`)

	doc, err := document.NewFromJSON(testJSONObj)
	if err != nil {
		t.Error(err)
		return
	}

	err = col.Save(doc)
	assert.NoError(t, err)

	// value check
	name, err := doc.Get("Name")
	assert.NoError(t, err)
	age, err := doc.Get("Age")
	assert.NoError(t, err)

	assert.Equal(t, "John", name)
	assert.Equal(t, int64(21), age)

	db.printDebugDB()
}

func TestDBUpdateDocWithFilter(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)
	col, err := newTestCollectionWithSchema(db)
	assert.NoError(t, err)

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21
	}`)

	doc, err := document.NewFromJSON(testJSONObj)
	if err != nil {
		t.Error(err)
		return
	}

	err = col.Save(doc)
	assert.NoError(t, err)

	_, err = col.UpdateWithFilter(`{Name: {_eq: "John"}}`, `{
		"Name": "Eric"
	}`)
	assert.NoError(t, err)

	doc, err = col.Get(doc.Key())
	assert.NoError(t, err)

	name, err := doc.Get("Name")
	assert.NoError(t, err)
	assert.Equal(t, "Eric", name)
}

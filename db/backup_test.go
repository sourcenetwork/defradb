// Copyright 2023 Democratized Data Foundation
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
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
)

func TestBasicExport_WithNormalFormatting_NoError(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}

	type Address {
		street: String
		city: String
	}`)
	require.NoError(t, err)
	col1, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc1, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`), col1.Definition())
	require.NoError(t, err)

	doc2, err := client.NewDocFromJSON([]byte(`{"name": "Bob", "age": 40}`), col1.Definition())
	require.NoError(t, err)

	err = col1.Create(ctx, doc1)
	require.NoError(t, err)

	err = col1.Create(ctx, doc2)
	require.NoError(t, err)

	col2, err := db.GetCollectionByName(ctx, "Address")
	require.NoError(t, err)

	doc3, err := client.NewDocFromJSON([]byte(`{"street": "101 Maple St", "city": "Toronto"}`), col2.Definition())
	require.NoError(t, err)

	err = col2.Create(ctx, doc3)
	require.NoError(t, err)

	txn, err := db.NewTxn(ctx, true)
	require.NoError(t, err)
	defer txn.Discard(ctx)

	ctx = SetContextIdentity(ctx, acpIdentity.None)
	ctx = SetContextTxn(ctx, txn)

	filepath := t.TempDir() + "/test.json"
	err = db.basicExport(ctx, &client.BackupConfig{Filepath: filepath})
	require.NoError(t, err)

	b, err := os.ReadFile(filepath)
	require.NoError(t, err)
	fileMap := map[string]any{}
	err = json.Unmarshal(b, &fileMap)
	require.NoError(t, err)

	expectedMap := map[string]any{}
	data := []byte(`{"Address":[{"_docID":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","_docIDNew":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","city":"Toronto","street":"101 Maple St"}],"User":[{"_docID":"bae-b94880d1-e6d2-542f-b9e0-5a369fafd0df","_docIDNew":"bae-b94880d1-e6d2-542f-b9e0-5a369fafd0df","age":40,"name":"Bob"},{"_docID":"bae-e933420a-988a-56f8-8952-6c245aebd519","_docIDNew":"bae-e933420a-988a-56f8-8952-6c245aebd519","age":30,"name":"John"}]}`)
	err = json.Unmarshal(data, &expectedMap)
	require.NoError(t, err)
	require.EqualValues(t, expectedMap, fileMap)
}

func TestBasicExport_WithPrettyFormatting_NoError(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}

	type Address {
		street: String
		city: String
	}`)
	require.NoError(t, err)

	col1, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc1, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`), col1.Definition())
	require.NoError(t, err)

	doc2, err := client.NewDocFromJSON([]byte(`{"name": "Bob", "age": 40}`), col1.Definition())
	require.NoError(t, err)

	err = col1.Create(ctx, doc1)
	require.NoError(t, err)

	err = col1.Create(ctx, doc2)
	require.NoError(t, err)

	col2, err := db.GetCollectionByName(ctx, "Address")
	require.NoError(t, err)

	doc3, err := client.NewDocFromJSON([]byte(`{"street": "101 Maple St", "city": "Toronto"}`), col2.Definition())
	require.NoError(t, err)

	err = col2.Create(ctx, doc3)
	require.NoError(t, err)

	txn, err := db.NewTxn(ctx, true)
	require.NoError(t, err)
	defer txn.Discard(ctx)

	ctx = SetContextIdentity(ctx, acpIdentity.None)
	ctx = SetContextTxn(ctx, txn)

	filepath := t.TempDir() + "/test.json"
	err = db.basicExport(ctx, &client.BackupConfig{Filepath: filepath, Pretty: true})
	require.NoError(t, err)

	b, err := os.ReadFile(filepath)
	require.NoError(t, err)
	fileMap := map[string]any{}
	err = json.Unmarshal(b, &fileMap)
	require.NoError(t, err)

	expectedMap := map[string]any{}
	data := []byte(`{"Address":[{"_docID":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","_docIDNew":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","city":"Toronto","street":"101 Maple St"}],"User":[{"_docID":"bae-b94880d1-e6d2-542f-b9e0-5a369fafd0df","_docIDNew":"bae-b94880d1-e6d2-542f-b9e0-5a369fafd0df","age":40,"name":"Bob"},{"_docID":"bae-e933420a-988a-56f8-8952-6c245aebd519","_docIDNew":"bae-e933420a-988a-56f8-8952-6c245aebd519","age":30,"name":"John"}]}`)
	err = json.Unmarshal(data, &expectedMap)
	require.NoError(t, err)
	require.EqualValues(t, expectedMap, fileMap)
}

func TestBasicExport_WithSingleCollection_NoError(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}

	type Address {
		street: String
		city: String
	}`)
	require.NoError(t, err)

	col1, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc1, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`), col1.Definition())
	require.NoError(t, err)

	doc2, err := client.NewDocFromJSON([]byte(`{"name": "Bob", "age": 40}`), col1.Definition())
	require.NoError(t, err)

	err = col1.Create(ctx, doc1)
	require.NoError(t, err)

	err = col1.Create(ctx, doc2)
	require.NoError(t, err)

	col2, err := db.GetCollectionByName(ctx, "Address")
	require.NoError(t, err)

	doc3, err := client.NewDocFromJSON([]byte(`{"street": "101 Maple St", "city": "Toronto"}`), col2.Definition())
	require.NoError(t, err)

	err = col2.Create(ctx, doc3)
	require.NoError(t, err)

	txn, err := db.NewTxn(ctx, true)
	require.NoError(t, err)
	defer txn.Discard(ctx)

	ctx = SetContextIdentity(ctx, acpIdentity.None)
	ctx = SetContextTxn(ctx, txn)

	filepath := t.TempDir() + "/test.json"
	err = db.basicExport(ctx, &client.BackupConfig{Filepath: filepath, Collections: []string{"Address"}})
	require.NoError(t, err)

	b, err := os.ReadFile(filepath)
	require.NoError(t, err)
	fileMap := map[string]any{}
	err = json.Unmarshal(b, &fileMap)
	require.NoError(t, err)

	expectedMap := map[string]any{}
	data := []byte(`{"Address":[{"_docID":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","_docIDNew":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","city":"Toronto","street":"101 Maple St"}]}`)
	err = json.Unmarshal(data, &expectedMap)
	require.NoError(t, err)
	require.EqualValues(t, expectedMap, fileMap)
}

func TestBasicExport_WithMultipleCollectionsAndUpdate_NoError(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddSchema(ctx, `type User {
		name: String
		age: Int
		book: [Book]
	}

	type Book {
		name: String
		author: User
	}`)
	require.NoError(t, err)

	col1, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc1, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`), col1.Definition())
	require.NoError(t, err)

	doc2, err := client.NewDocFromJSON([]byte(`{"name": "Bob", "age": 31}`), col1.Definition())
	require.NoError(t, err)

	err = col1.Create(ctx, doc1)
	require.NoError(t, err)

	err = col1.Create(ctx, doc2)
	require.NoError(t, err)

	col2, err := db.GetCollectionByName(ctx, "Book")
	require.NoError(t, err)

	doc3, err := client.NewDocFromJSON([]byte(`{"name": "John and the sourcerers' stone", "author": "bae-e933420a-988a-56f8-8952-6c245aebd519"}`), col2.Definition())
	require.NoError(t, err)

	doc4, err := client.NewDocFromJSON([]byte(`{"name": "Game of chains", "author": "bae-e933420a-988a-56f8-8952-6c245aebd519"}`), col2.Definition())
	require.NoError(t, err)

	err = col2.Create(ctx, doc3)
	require.NoError(t, err)
	err = col2.Create(ctx, doc4)
	require.NoError(t, err)

	err = doc1.Set("age", 31)
	require.NoError(t, err)

	err = col1.Update(ctx, doc1)
	require.NoError(t, err)

	txn, err := db.NewTxn(ctx, true)
	require.NoError(t, err)
	defer txn.Discard(ctx)

	ctx = SetContextIdentity(ctx, acpIdentity.None)
	ctx = SetContextTxn(ctx, txn)

	filepath := t.TempDir() + "/test.json"
	err = db.basicExport(ctx, &client.BackupConfig{Filepath: filepath})
	require.NoError(t, err)

	b, err := os.ReadFile(filepath)
	require.NoError(t, err)
	fileMap := map[string]any{}
	err = json.Unmarshal(b, &fileMap)
	require.NoError(t, err)

	expectedMap := map[string]any{}
	data := []byte(`{"Book":[{"_docID":"bae-4399f189-138d-5d49-9e25-82e78463677b","_docIDNew":"bae-78a40f28-a4b8-5dca-be44-392b0f96d0ff","author_id":"bae-807ea028-6c13-5f86-a72b-46e8b715a162","name":"Game of chains"},{"_docID":"bae-5cf2fec3-d8ed-50d5-8286-39109853d2da","_docIDNew":"bae-edeade01-2d21-5d6d-aadf-efc5a5279de5","author_id":"bae-807ea028-6c13-5f86-a72b-46e8b715a162","name":"John and the sourcerers' stone"}],"User":[{"_docID":"bae-0648f44e-74e8-593b-a662-3310ec278927","_docIDNew":"bae-0648f44e-74e8-593b-a662-3310ec278927","age":31,"name":"Bob"},{"_docID":"bae-e933420a-988a-56f8-8952-6c245aebd519","_docIDNew":"bae-807ea028-6c13-5f86-a72b-46e8b715a162","age":31,"name":"John"}]}`)
	err = json.Unmarshal(data, &expectedMap)
	require.NoError(t, err)
	require.EqualValues(t, expectedMap, fileMap)
}

func TestBasicExport_EnsureFileOverwrite_NoError(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}

	type Address {
		street: String
		city: String
	}`)
	require.NoError(t, err)

	col1, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc1, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`), col1.Definition())
	require.NoError(t, err)

	doc2, err := client.NewDocFromJSON([]byte(`{"name": "Bob", "age": 40}`), col1.Definition())
	require.NoError(t, err)

	err = col1.Create(ctx, doc1)
	require.NoError(t, err)

	err = col1.Create(ctx, doc2)
	require.NoError(t, err)

	col2, err := db.GetCollectionByName(ctx, "Address")
	require.NoError(t, err)

	doc3, err := client.NewDocFromJSON([]byte(`{"street": "101 Maple St", "city": "Toronto"}`), col2.Definition())
	require.NoError(t, err)

	err = col2.Create(ctx, doc3)
	require.NoError(t, err)

	txn, err := db.NewTxn(ctx, true)
	require.NoError(t, err)
	defer txn.Discard(ctx)

	ctx = SetContextIdentity(ctx, acpIdentity.None)
	ctx = SetContextTxn(ctx, txn)

	filepath := t.TempDir() + "/test.json"

	err = os.WriteFile(
		filepath,
		[]byte(`{"Address":[{"_docID":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","_docIDNew":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","city":"Toronto","street":"101 Maple St"}],"User":[{"_docID":"bae-b94880d1-e6d2-542f-b9e0-5a369fafd0df","_docIDNew":"bae-b94880d1-e6d2-542f-b9e0-5a369fafd0df","age":40,"name":"Bob"},{"_docID":"bae-e933420a-988a-56f8-8952-6c245aebd519","_docIDNew":"bae-e933420a-988a-56f8-8952-6c245aebd519","age":30,"name":"John"}]}`),
		0664,
	)
	require.NoError(t, err)

	err = db.basicExport(ctx, &client.BackupConfig{Filepath: filepath, Collections: []string{"Address"}})
	require.NoError(t, err)

	b, err := os.ReadFile(filepath)
	require.NoError(t, err)
	fileMap := map[string]any{}
	err = json.Unmarshal(b, &fileMap)
	require.NoError(t, err)

	expectedMap := map[string]any{}
	data := []byte(`{"Address":[{"_docID":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","_docIDNew":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","city":"Toronto","street":"101 Maple St"}]}`)
	err = json.Unmarshal(data, &expectedMap)
	require.NoError(t, err)
	require.EqualValues(t, expectedMap, fileMap)
}

func TestBasicImport_WithMultipleCollectionsAndObjects_NoError(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}

	type Address {
		street: String
		city: String
	}`)
	require.NoError(t, err)

	txn, err := db.NewTxn(ctx, false)
	require.NoError(t, err)

	ctx = SetContextIdentity(ctx, acpIdentity.None)
	ctx = SetContextTxn(ctx, txn)

	filepath := t.TempDir() + "/test.json"

	err = os.WriteFile(
		filepath,
		[]byte(`{"Address":[{"_docID":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","_docIDNew":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","city":"Toronto","street":"101 Maple St"}],"User":[{"_docID":"bae-b94880d1-e6d2-542f-b9e0-5a369fafd0df","_docIDNew":"bae-b94880d1-e6d2-542f-b9e0-5a369fafd0df","age":40,"name":"Bob"},{"_docID":"bae-e933420a-988a-56f8-8952-6c245aebd519","_docIDNew":"bae-e933420a-988a-56f8-8952-6c245aebd519","age":30,"name":"John"}]}`),
		0664,
	)
	require.NoError(t, err)

	err = db.basicImport(ctx, filepath)
	require.NoError(t, err)
	err = txn.Commit(ctx)
	require.NoError(t, err)

	txn, err = db.NewTxn(ctx, true)
	require.NoError(t, err)

	ctx = SetContextIdentity(ctx, acpIdentity.None)
	ctx = SetContextTxn(ctx, txn)

	col1, err := db.getCollectionByName(ctx, "Address")
	require.NoError(t, err)

	key1, err := client.NewDocIDFromString("bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f")
	require.NoError(t, err)
	_, err = col1.Get(ctx, key1, false)
	require.NoError(t, err)

	col2, err := db.getCollectionByName(ctx, "User")
	require.NoError(t, err)

	key2, err := client.NewDocIDFromString("bae-b94880d1-e6d2-542f-b9e0-5a369fafd0df")
	require.NoError(t, err)
	_, err = col2.Get(ctx, key2, false)
	require.NoError(t, err)

	key3, err := client.NewDocIDFromString("bae-e933420a-988a-56f8-8952-6c245aebd519")
	require.NoError(t, err)
	_, err = col2.Get(ctx, key3, false)
	require.NoError(t, err)
}

func TestBasicImport_WithJSONArray_ReturnError(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}

	type Address {
		street: String
		city: String
	}`)
	require.NoError(t, err)

	txn, err := db.NewTxn(ctx, false)
	require.NoError(t, err)
	ctx = SetContextTxn(ctx, txn)

	filepath := t.TempDir() + "/test.json"

	err = os.WriteFile(
		filepath,
		[]byte(`["Address":[{"_docID":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","_docIDNew":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","city":"Toronto","street":"101 Maple St"}],"User":[{"_docID":"bae-b94880d1-e6d2-542f-b9e0-5a369fafd0df","_docIDNew":"bae-b94880d1-e6d2-542f-b9e0-5a369fafd0df","age":40,"name":"Bob"},{"_docID":"bae-e933420a-988a-56f8-8952-6c245aebd519","_docIDNew":"bae-e933420a-988a-56f8-8952-6c245aebd519","age":30,"name":"John"}]]`),
		0664,
	)
	require.NoError(t, err)

	err = db.basicImport(ctx, filepath)
	require.ErrorIs(t, err, ErrExpectedJSONObject)
	err = txn.Commit(ctx)
	require.NoError(t, err)
}

func TestBasicImport_WithObjectCollection_ReturnError(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}

	type Address {
		street: String
		city: String
	}`)
	require.NoError(t, err)

	txn, err := db.NewTxn(ctx, false)
	require.NoError(t, err)
	ctx = SetContextTxn(ctx, txn)

	filepath := t.TempDir() + "/test.json"

	err = os.WriteFile(
		filepath,
		[]byte(`{"Address":{"_docID":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","_docIDNew":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","city":"Toronto","street":"101 Maple St"}}`),
		0664,
	)
	require.NoError(t, err)

	err = db.basicImport(ctx, filepath)
	require.ErrorIs(t, err, ErrExpectedJSONArray)
	err = txn.Commit(ctx)
	require.NoError(t, err)
}

func TestBasicImport_WithInvalidFilepath_ReturnError(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}

	type Address {
		street: String
		city: String
	}`)
	require.NoError(t, err)

	txn, err := db.NewTxn(ctx, false)
	require.NoError(t, err)
	ctx = SetContextTxn(ctx, txn)

	filepath := t.TempDir() + "/test.json"

	err = os.WriteFile(
		filepath,
		[]byte(`{"Address":{"_docID":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","_docIDNew":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","city":"Toronto","street":"101 Maple St"}}`),
		0664,
	)
	require.NoError(t, err)

	wrongFilepath := t.TempDir() + "/some/test.json"
	err = db.basicImport(ctx, wrongFilepath)
	require.ErrorIs(t, err, os.ErrNotExist)
	err = txn.Commit(ctx)
	require.NoError(t, err)
}

func TestBasicImport_WithInvalidCollection_ReturnError(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}

	type Address {
		street: String
		city: String
	}`)
	require.NoError(t, err)

	txn, err := db.NewTxn(ctx, false)
	require.NoError(t, err)
	ctx = SetContextTxn(ctx, txn)

	filepath := t.TempDir() + "/test.json"

	err = os.WriteFile(
		filepath,
		[]byte(`{"Addresses":{"_docID":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","_docIDNew":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","city":"Toronto","street":"101 Maple St"}}`),
		0664,
	)
	require.NoError(t, err)

	err = db.basicImport(ctx, filepath)
	require.ErrorIs(t, err, ErrFailedToGetCollection)
	err = txn.Commit(ctx)
	require.NoError(t, err)
}

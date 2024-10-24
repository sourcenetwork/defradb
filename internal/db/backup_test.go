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

	"github.com/sourcenetwork/defradb/acp/identity"
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

	ctx = identity.WithContext(ctx, acpIdentity.None)
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
	data := []byte(`{"Address":[{"_docID":"bae-41e1a410-df86-5846-939e-4470a8d8cb0c","_docIDNew":"bae-41e1a410-df86-5846-939e-4470a8d8cb0c","city":"Toronto","street":"101 Maple St"}],"User":[{"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","_docIDNew":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","age":30,"name":"John"},{"_docID":"bae-f8a0f1e4-129e-50ab-98ed-1aa110810fb2","_docIDNew":"bae-f8a0f1e4-129e-50ab-98ed-1aa110810fb2","age":40,"name":"Bob"}]}`)
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

	ctx = identity.WithContext(ctx, acpIdentity.None)
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
	data := []byte(`{"Address":[{"_docID":"bae-41e1a410-df86-5846-939e-4470a8d8cb0c","_docIDNew":"bae-41e1a410-df86-5846-939e-4470a8d8cb0c","city":"Toronto","street":"101 Maple St"}],"User":[{"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","_docIDNew":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","age":30,"name":"John"},{"_docID":"bae-f8a0f1e4-129e-50ab-98ed-1aa110810fb2","_docIDNew":"bae-f8a0f1e4-129e-50ab-98ed-1aa110810fb2","age":40,"name":"Bob"}]}`)
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

	ctx = identity.WithContext(ctx, acpIdentity.None)
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
	data := []byte(`{"Address":[{"_docID":"bae-41e1a410-df86-5846-939e-4470a8d8cb0c","_docIDNew":"bae-41e1a410-df86-5846-939e-4470a8d8cb0c","city":"Toronto","street":"101 Maple St"}]}`)
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

	doc3, err := client.NewDocFromJSON([]byte(`{"name": "John and the sourcerers' stone", "author": "bae-7fca96a2-5f01-5558-a81f-09b47587f26d"}`), col2.Definition())
	require.NoError(t, err)

	doc4, err := client.NewDocFromJSON([]byte(`{"name": "Game of chains", "author": "bae-7fca96a2-5f01-5558-a81f-09b47587f26d"}`), col2.Definition())
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

	ctx = identity.WithContext(ctx, acpIdentity.None)
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
	data := []byte(`{"Book":[{"_docID":"bae-4a28c746-ccbf-5511-91a9-391036f42f80", "_docIDNew":"bae-d821f684-47de-5b63-b9c7-6eccec368e52", "author_id":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f", "name":"Game of chains"}, {"_docID":"bae-8c8be5c6-d26b-50d4-9378-2acd5fe6959d", "_docIDNew":"bae-c94e52f8-6e91-522c-b6a6-38346a06b3d2", "author_id":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f", "name":"John and the sourcerers' stone"}], "User":[{"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d", "_docIDNew":"bae-9918e1ec-c62b-5de2-8fbf-c82795b8ac7f", "age":31, "name":"John"}, {"_docID":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f", "_docIDNew":"bae-ebfe11e2-045d-525d-9fb7-2abb961dc84f", "age":31, "name":"Bob"}]}`)
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

	ctx = identity.WithContext(ctx, acpIdentity.None)
	ctx = SetContextTxn(ctx, txn)

	filepath := t.TempDir() + "/test.json"

	err = os.WriteFile(
		filepath,
		[]byte(`{"Address":[{"_docID":"bae-41e1a410-df86-5846-939e-4470a8d8cb0c","_docIDNew":"bae-41e1a410-df86-5846-939e-4470a8d8cb0c","city":"Toronto","street":"101 Maple St"}],"User":[{"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","_docIDNew":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","age":40,"name":"Bob"},{"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","_docIDNew":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","age":30,"name":"John"}]}`),
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
	data := []byte(`{"Address":[{"_docID":"bae-41e1a410-df86-5846-939e-4470a8d8cb0c","_docIDNew":"bae-41e1a410-df86-5846-939e-4470a8d8cb0c","city":"Toronto","street":"101 Maple St"}]}`)
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

	ctx = identity.WithContext(ctx, acpIdentity.None)
	ctx = SetContextTxn(ctx, txn)

	filepath := t.TempDir() + "/test.json"

	err = os.WriteFile(
		filepath,
		[]byte(`{"Address":[{"_docID":"bae-41e1a410-df86-5846-939e-4470a8d8cb0c","_docIDNew":"bae-41e1a410-df86-5846-939e-4470a8d8cb0c","city":"Toronto","street":"101 Maple St"}],"User":[{"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","_docIDNew":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","age":40,"name":"Bob"},{"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","_docIDNew":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","age":30,"name":"John"}]}`),
		0664,
	)
	require.NoError(t, err)

	err = db.basicImport(ctx, filepath)
	require.NoError(t, err)
	err = txn.Commit(ctx)
	require.NoError(t, err)

	txn, err = db.NewTxn(ctx, true)
	require.NoError(t, err)

	ctx = identity.WithContext(ctx, acpIdentity.None)
	ctx = SetContextTxn(ctx, txn)

	col1, err := db.getCollectionByName(ctx, "Address")
	require.NoError(t, err)

	key1, err := client.NewDocIDFromString("bae-41e1a410-df86-5846-939e-4470a8d8cb0c")
	require.NoError(t, err)
	_, err = col1.Get(ctx, key1, false)
	require.NoError(t, err)

	col2, err := db.getCollectionByName(ctx, "User")
	require.NoError(t, err)

	key2, err := client.NewDocIDFromString("bae-7fca96a2-5f01-5558-a81f-09b47587f26d")
	require.NoError(t, err)
	_, err = col2.Get(ctx, key2, false)
	require.NoError(t, err)

	key3, err := client.NewDocIDFromString("bae-7fca96a2-5f01-5558-a81f-09b47587f26d")
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
		[]byte(`["Address":[{"_docID":"bae-41e1a410-df86-5846-939e-4470a8d8cb0c","_docIDNew":"bae-41e1a410-df86-5846-939e-4470a8d8cb0c","city":"Toronto","street":"101 Maple St"}],"User":[{"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","_docIDNew":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","age":40,"name":"Bob"},{"_docID":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","_docIDNew":"bae-7fca96a2-5f01-5558-a81f-09b47587f26d","age":30,"name":"John"}]]`),
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
		[]byte(`{"Address":{"_docID":"bae-41e1a410-df86-5846-939e-4470a8d8cb0c","_docIDNew":"bae-41e1a410-df86-5846-939e-4470a8d8cb0c","city":"Toronto","street":"101 Maple St"}}`),
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
		[]byte(`{"Address":{"_docID":"bae-41e1a410-df86-5846-939e-4470a8d8cb0c","_docIDNew":"bae-41e1a410-df86-5846-939e-4470a8d8cb0c","city":"Toronto","street":"101 Maple St"}}`),
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
		[]byte(`{"Addresses":{"_docID":"bae-41e1a410-df86-5846-939e-4470a8d8cb0c","_docIDNew":"bae-41e1a410-df86-5846-939e-4470a8d8cb0c","city":"Toronto","street":"101 Maple St"}}`),
		0664,
	)
	require.NoError(t, err)

	err = db.basicImport(ctx, filepath)
	require.ErrorIs(t, err, ErrFailedToGetCollection)
	err = txn.Commit(ctx)
	require.NoError(t, err)
}

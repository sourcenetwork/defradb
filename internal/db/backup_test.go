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
	db, err := newBadgerDB(ctx)
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
	ctx = InitContext(ctx, txn)

	filepath := t.TempDir() + "/test.json"
	err = db.basicExport(ctx, &client.BackupConfig{Filepath: filepath})
	require.NoError(t, err)

	b, err := os.ReadFile(filepath)
	require.NoError(t, err)
	fileMap := map[string]any{}
	err = json.Unmarshal(b, &fileMap)
	require.NoError(t, err)

	expectedMap := map[string]any{}
	data := []byte(`{"User":[{"_docID":"bae-74242e1b-614c-5007-af7d-9c25c1d5b1a9","_docIDNew":"bae-74242e1b-614c-5007-af7d-9c25c1d5b1a9","age":40,"name":"Bob"},{"_docID":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","_docIDNew":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","age":30,"name":"John"}],"Address":[{"_docID":"bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa","_docIDNew":"bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa","city":"Toronto","street":"101 Maple St"}]}`)
	err = json.Unmarshal(data, &expectedMap)
	require.NoError(t, err)
	require.EqualValues(t, expectedMap, fileMap)
}

func TestBasicExport_WithPrettyFormatting_NoError(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
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
	ctx = InitContext(ctx, txn)

	filepath := t.TempDir() + "/test.json"
	err = db.basicExport(ctx, &client.BackupConfig{Filepath: filepath, Pretty: true})
	require.NoError(t, err)

	b, err := os.ReadFile(filepath)
	require.NoError(t, err)
	fileMap := map[string]any{}
	err = json.Unmarshal(b, &fileMap)
	require.NoError(t, err)

	expectedMap := map[string]any{}
	data := []byte(`{"User": [{"_docID": "bae-74242e1b-614c-5007-af7d-9c25c1d5b1a9","_docIDNew": "bae-74242e1b-614c-5007-af7d-9c25c1d5b1a9","age": 40,"name": "Bob"},{"_docID": "bae-a911f9cc-217a-58a3-a2f4-96548197403e","_docIDNew": "bae-a911f9cc-217a-58a3-a2f4-96548197403e","age": 30,"name": "John"}],"Address": [{"_docID": "bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa","_docIDNew": "bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa","city": "Toronto","street": "101 Maple St"}]}`)
	err = json.Unmarshal(data, &expectedMap)
	require.NoError(t, err)
	require.EqualValues(t, expectedMap, fileMap)
}

func TestBasicExport_WithSingleCollection_NoError(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
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
	ctx = InitContext(ctx, txn)

	filepath := t.TempDir() + "/test.json"
	err = db.basicExport(ctx, &client.BackupConfig{Filepath: filepath, Collections: []string{"Address"}})
	require.NoError(t, err)

	b, err := os.ReadFile(filepath)
	require.NoError(t, err)
	fileMap := map[string]any{}
	err = json.Unmarshal(b, &fileMap)
	require.NoError(t, err)

	expectedMap := map[string]any{}
	data := []byte(`{"Address":[{"_docID":"bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa","_docIDNew":"bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa","city":"Toronto","street":"101 Maple St"}]}`)
	err = json.Unmarshal(data, &expectedMap)
	require.NoError(t, err)
	require.EqualValues(t, expectedMap, fileMap)
}

func TestBasicExport_WithMultipleCollectionsAndUpdate_NoError(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
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

	doc3, err := client.NewDocFromJSON([]byte(`{"name": "John and the sourcerers' stone", "author": "bae-a911f9cc-217a-58a3-a2f4-96548197403e"}`), col2.Definition())
	require.NoError(t, err)

	doc4, err := client.NewDocFromJSON([]byte(`{"name": "Game of chains", "author": "bae-a911f9cc-217a-58a3-a2f4-96548197403e"}`), col2.Definition())
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
	ctx = InitContext(ctx, txn)

	filepath := t.TempDir() + "/test.json"
	err = db.basicExport(ctx, &client.BackupConfig{Filepath: filepath})
	require.NoError(t, err)

	b, err := os.ReadFile(filepath)
	require.NoError(t, err)
	fileMap := map[string]any{}
	err = json.Unmarshal(b, &fileMap)
	require.NoError(t, err)

	expectedMap := map[string]any{}
	data := []byte(`{"User":[{"_docID":"bae-88fea952-a678-5e05-9895-8a86ac6abc3b","_docIDNew":"bae-88fea952-a678-5e05-9895-8a86ac6abc3b","age":31,"name":"Bob"},{"_docID":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","_docIDNew":"bae-97f27fca-8b97-59f1-afa1-2e63140de933","age":31,"name":"John"}],"Book":[{"_docID":"bae-191238ef-acd2-5d9f-8d95-dcc15415fc75","_docIDNew":"bae-b5ea3d12-0519-5a5f-a7bc-9f5fee62c12f","author_id":"bae-97f27fca-8b97-59f1-afa1-2e63140de933","name":"Game of chains"},{"_docID":"bae-f97cb90a-20db-5595-b193-89bdf50bdee8","_docIDNew":"bae-8a319b41-e061-5d19-a847-388fa51f732c","author_id":"bae-97f27fca-8b97-59f1-afa1-2e63140de933","name":"John and the sourcerers' stone"}]}`)
	err = json.Unmarshal(data, &expectedMap)
	require.NoError(t, err)
	require.EqualValues(t, expectedMap, fileMap)
}

func TestBasicExport_EnsureFileOverwrite_NoError(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
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
	ctx = InitContext(ctx, txn)

	filepath := t.TempDir() + "/test.json"

	err = os.WriteFile(
		filepath,
		[]byte(`{"Address":[{"_docID":"bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa","_docIDNew":"bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa","city":"Toronto","street":"101 Maple St"}],"User":[{"_docID":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","_docIDNew":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","age":40,"name":"Bob"},{"_docID":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","_docIDNew":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","age":30,"name":"John"}]}`),
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
	data := []byte(`{"Address":[{"_docID":"bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa","_docIDNew":"bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa","city":"Toronto","street":"101 Maple St"}]}`)
	err = json.Unmarshal(data, &expectedMap)
	require.NoError(t, err)
	require.EqualValues(t, expectedMap, fileMap)
}

func TestBasicImport_WithMultipleCollectionsAndObjects_NoError(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
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
	ctx = InitContext(ctx, txn)

	filepath := t.TempDir() + "/test.json"

	err = os.WriteFile(
		filepath,
		[]byte(`{"Address":[{"_docID":"bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa","_docIDNew":"bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa","city":"Toronto","street":"101 Maple St"}],"User":[{"_docID":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","_docIDNew":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","age":40,"name":"Bob"},{"_docID":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","_docIDNew":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","age":30,"name":"John"}]}`),
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
	ctx = InitContext(ctx, txn)

	col1, err := db.getCollectionByName(ctx, "Address")
	require.NoError(t, err)

	key1, err := client.NewDocIDFromString("bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa")
	require.NoError(t, err)
	_, err = col1.Get(ctx, key1, false)
	require.NoError(t, err)

	col2, err := db.getCollectionByName(ctx, "User")
	require.NoError(t, err)

	key2, err := client.NewDocIDFromString("bae-a911f9cc-217a-58a3-a2f4-96548197403e")
	require.NoError(t, err)
	_, err = col2.Get(ctx, key2, false)
	require.NoError(t, err)

	key3, err := client.NewDocIDFromString("bae-a911f9cc-217a-58a3-a2f4-96548197403e")
	require.NoError(t, err)
	_, err = col2.Get(ctx, key3, false)
	require.NoError(t, err)
}

func TestBasicImport_WithJSONArray_ReturnError(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
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
	ctx = InitContext(ctx, txn)

	filepath := t.TempDir() + "/test.json"

	err = os.WriteFile(
		filepath,
		[]byte(`["Address":[{"_docID":"bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa","_docIDNew":"bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa","city":"Toronto","street":"101 Maple St"}],"User":[{"_docID":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","_docIDNew":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","age":40,"name":"Bob"},{"_docID":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","_docIDNew":"bae-a911f9cc-217a-58a3-a2f4-96548197403e","age":30,"name":"John"}]]`),
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
	db, err := newBadgerDB(ctx)
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
	ctx = InitContext(ctx, txn)

	filepath := t.TempDir() + "/test.json"

	err = os.WriteFile(
		filepath,
		[]byte(`{"Address":{"_docID":"bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa","_docIDNew":"bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa","city":"Toronto","street":"101 Maple St"}}`),
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
	db, err := newBadgerDB(ctx)
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
	ctx = InitContext(ctx, txn)

	filepath := t.TempDir() + "/test.json"

	err = os.WriteFile(
		filepath,
		[]byte(`{"Address":{"_docID":"bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa","_docIDNew":"bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa","city":"Toronto","street":"101 Maple St"}}`),
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
	db, err := newBadgerDB(ctx)
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
	ctx = InitContext(ctx, txn)

	filepath := t.TempDir() + "/test.json"

	err = os.WriteFile(
		filepath,
		[]byte(`{"Addresses":{"_docID":"bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa","_docIDNew":"bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa","city":"Toronto","street":"101 Maple St"}}`),
		0664,
	)
	require.NoError(t, err)

	err = db.basicImport(ctx, filepath)
	require.ErrorIs(t, err, ErrFailedToGetCollection)
	err = txn.Commit(ctx)
	require.NoError(t, err)
}

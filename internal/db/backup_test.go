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

	doc1, err := client.NewDocFromJSON(ctx, []byte(`{"name": "John", "age": 30}`), col1.Version())
	require.NoError(t, err)

	doc2, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Bob", "age": 40}`), col1.Version())
	require.NoError(t, err)

	err = col1.Create(ctx, doc1)
	require.NoError(t, err)

	err = col1.Create(ctx, doc2)
	require.NoError(t, err)

	col2, err := db.GetCollectionByName(ctx, "Address")
	require.NoError(t, err)

	doc3, err := client.NewDocFromJSON(ctx, []byte(`{"street": "101 Maple St", "city": "Toronto"}`), col2.Version())
	require.NoError(t, err)

	err = col2.Create(ctx, doc3)
	require.NoError(t, err)

	txn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer txn.Discard()

	ctx = identity.WithContext(ctx, identity.None)
	ctx = InitContext(ctx, txn)

	filepath := t.TempDir() + "/test.json"
	err = db.basicExport(ctx, &client.BackupConfig{Filepath: filepath})
	require.NoError(t, err)

	b, err := os.ReadFile(filepath)
	require.NoError(t, err)
	fileMap := map[string]any{}
	err = json.Unmarshal(b, &fileMap)
	require.NoError(t, err)

	// Verify structure instead of exact docIDs
	require.Contains(t, fileMap, "User")
	require.Contains(t, fileMap, "Address")

	users, _ := fileMap["User"].([]any)
	require.Len(t, users, 2)

	addresses, _ := fileMap["Address"].([]any)
	require.Len(t, addresses, 1)

	// Verify User documents contain expected data (order may vary)
	userNames := make([]string, 2)
	userAges := make([]float64, 2)
	for i, u := range users {
		user, _ := u.(map[string]any)
		require.Contains(t, user, "_docID")
		require.Contains(t, user, "_docIDNew")
		require.Contains(t, user, "name")
		require.Contains(t, user, "age")
		userNames[i], _ = user["name"].(string)
		userAges[i], _ = user["age"].(float64)
	}
	require.ElementsMatch(t, []string{"John", "Bob"}, userNames)
	require.ElementsMatch(t, []float64{30, 40}, userAges)

	// Verify Address document
	address, _ := addresses[0].(map[string]any)
	require.Contains(t, address, "_docID")
	require.Contains(t, address, "_docIDNew")
	require.Equal(t, "Toronto", address["city"])
	require.Equal(t, "101 Maple St", address["street"])
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

	doc1, err := client.NewDocFromJSON(ctx, []byte(`{"name": "John", "age": 30}`), col1.Version())
	require.NoError(t, err)

	doc2, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Bob", "age": 40}`), col1.Version())
	require.NoError(t, err)

	err = col1.Create(ctx, doc1)
	require.NoError(t, err)

	err = col1.Create(ctx, doc2)
	require.NoError(t, err)

	col2, err := db.GetCollectionByName(ctx, "Address")
	require.NoError(t, err)

	doc3, err := client.NewDocFromJSON(ctx, []byte(`{"street": "101 Maple St", "city": "Toronto"}`), col2.Version())
	require.NoError(t, err)

	err = col2.Create(ctx, doc3)
	require.NoError(t, err)

	txn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer txn.Discard()

	ctx = identity.WithContext(ctx, identity.None)
	ctx = InitContext(ctx, txn)

	filepath := t.TempDir() + "/test.json"
	err = db.basicExport(ctx, &client.BackupConfig{Filepath: filepath, Pretty: true})
	require.NoError(t, err)

	b, err := os.ReadFile(filepath)
	require.NoError(t, err)
	fileMap := map[string]any{}
	err = json.Unmarshal(b, &fileMap)
	require.NoError(t, err)

	// Verify structure instead of exact docIDs
	require.Contains(t, fileMap, "User")
	require.Contains(t, fileMap, "Address")

	users, _ := fileMap["User"].([]any)
	require.Len(t, users, 2)

	addresses, _ := fileMap["Address"].([]any)
	require.Len(t, addresses, 1)

	// Verify User documents contain expected data (order may vary)
	userNames := make([]string, 2)
	userAges := make([]float64, 2)
	for i, u := range users {
		user, _ := u.(map[string]any)
		require.Contains(t, user, "_docID")
		require.Contains(t, user, "_docIDNew")
		require.Contains(t, user, "name")
		require.Contains(t, user, "age")
		userNames[i], _ = user["name"].(string)
		userAges[i], _ = user["age"].(float64)
	}
	require.ElementsMatch(t, []string{"John", "Bob"}, userNames)
	require.ElementsMatch(t, []float64{30, 40}, userAges)

	// Verify Address document
	address, _ := addresses[0].(map[string]any)
	require.Contains(t, address, "_docID")
	require.Contains(t, address, "_docIDNew")
	require.Equal(t, "Toronto", address["city"])
	require.Equal(t, "101 Maple St", address["street"])
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

	doc1, err := client.NewDocFromJSON(ctx, []byte(`{"name": "John", "age": 30}`), col1.Version())
	require.NoError(t, err)

	doc2, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Bob", "age": 40}`), col1.Version())
	require.NoError(t, err)

	err = col1.Create(ctx, doc1)
	require.NoError(t, err)

	err = col1.Create(ctx, doc2)
	require.NoError(t, err)

	col2, err := db.GetCollectionByName(ctx, "Address")
	require.NoError(t, err)

	doc3, err := client.NewDocFromJSON(ctx, []byte(`{"street": "101 Maple St", "city": "Toronto"}`), col2.Version())
	require.NoError(t, err)

	err = col2.Create(ctx, doc3)
	require.NoError(t, err)

	txn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer txn.Discard()

	ctx = identity.WithContext(ctx, identity.None)
	ctx = InitContext(ctx, txn)

	filepath := t.TempDir() + "/test.json"
	err = db.basicExport(ctx, &client.BackupConfig{Filepath: filepath, Collections: []string{"Address"}})
	require.NoError(t, err)

	b, err := os.ReadFile(filepath)
	require.NoError(t, err)
	fileMap := map[string]any{}
	err = json.Unmarshal(b, &fileMap)
	require.NoError(t, err)

	// Verify structure instead of exact docIDs
	require.Contains(t, fileMap, "Address")
	require.NotContains(t, fileMap, "User") // Should only have Address collection

	addresses, _ := fileMap["Address"].([]any)
	require.Len(t, addresses, 1)

	// Verify Address document
	address, _ := addresses[0].(map[string]any)
	require.Contains(t, address, "_docID")
	require.Contains(t, address, "_docIDNew")
	require.Equal(t, "Toronto", address["city"])
	require.Equal(t, "101 Maple St", address["street"])
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

	doc1, err := client.NewDocFromJSON(ctx, []byte(`{"name": "John", "age": 30}`), col1.Version())
	require.NoError(t, err)

	doc2, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Bob", "age": 31}`), col1.Version())
	require.NoError(t, err)

	err = col1.Create(ctx, doc1)
	require.NoError(t, err)

	err = col1.Create(ctx, doc2)
	require.NoError(t, err)

	col2, err := db.GetCollectionByName(ctx, "Book")
	require.NoError(t, err)

	// Use the actual doc1 ID for the relationship
	doc1ID := doc1.ID().String()
	doc3, err := client.NewDocFromJSON(ctx, []byte(`{"name": "John and the sourcerers' stone", "author": "`+doc1ID+`"}`), col2.Version())
	require.NoError(t, err)

	doc4, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Game of chains", "author": "`+doc1ID+`"}`), col2.Version())
	require.NoError(t, err)

	err = col2.Create(ctx, doc3)
	require.NoError(t, err)
	err = col2.Create(ctx, doc4)
	require.NoError(t, err)

	err = doc1.Set(ctx, "age", 31)
	require.NoError(t, err)

	err = col1.Update(ctx, doc1)
	require.NoError(t, err)

	txn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer txn.Discard()

	ctx = identity.WithContext(ctx, identity.None)
	ctx = InitContext(ctx, txn)

	filepath := t.TempDir() + "/test.json"
	err = db.basicExport(ctx, &client.BackupConfig{Filepath: filepath})
	require.NoError(t, err)

	b, err := os.ReadFile(filepath)
	require.NoError(t, err)
	fileMap := map[string]any{}
	err = json.Unmarshal(b, &fileMap)
	require.NoError(t, err)

	// Verify structure instead of exact docIDs
	require.Contains(t, fileMap, "User")
	require.Contains(t, fileMap, "Book")

	users, _ := fileMap["User"].([]any)
	require.Len(t, users, 2)

	books, _ := fileMap["Book"].([]any)
	require.Len(t, books, 2)

	// Get the new docID for John after update
	var johnNewDocID string
	for _, u := range users {
		user, _ := u.(map[string]any)
		switch user["name"] {
		case "John":
			require.Equal(t, float64(31), user["age"])
			johnNewDocID, _ = user["_docIDNew"].(string)
		case "Bob":
			require.Equal(t, float64(31), user["age"])
		}
	}

	// Verify both books reference the correct author
	bookNames := make([]string, 2)
	for i, b := range books {
		book, _ := b.(map[string]any)
		require.Contains(t, book, "author_id")
		require.Equal(t, johnNewDocID, book["author_id"])
		bookNames[i], _ = book["name"].(string)
	}
	require.ElementsMatch(t, []string{"John and the sourcerers' stone", "Game of chains"}, bookNames)
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

	doc1, err := client.NewDocFromJSON(ctx, []byte(`{"name": "John", "age": 30}`), col1.Version())
	require.NoError(t, err)

	doc2, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Bob", "age": 40}`), col1.Version())
	require.NoError(t, err)

	err = col1.Create(ctx, doc1)
	require.NoError(t, err)

	err = col1.Create(ctx, doc2)
	require.NoError(t, err)

	col2, err := db.GetCollectionByName(ctx, "Address")
	require.NoError(t, err)

	doc3, err := client.NewDocFromJSON(ctx, []byte(`{"street": "101 Maple St", "city": "Toronto"}`), col2.Version())
	require.NoError(t, err)

	err = col2.Create(ctx, doc3)
	require.NoError(t, err)

	txn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer txn.Discard()

	ctx = identity.WithContext(ctx, identity.None)
	ctx = InitContext(ctx, txn)

	filepath := t.TempDir() + "/test.json"

	err = os.WriteFile(
		filepath,
		[]byte(`{"Address":[{"_docID":"bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa","_docIDNew":"bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa","city":"Toronto","street":"101 Maple St"}],"User":[{"_docID":"bae-3fc941b7-505c-5ce2-91a0-b180930ec8a9","_docIDNew":"bae-3fc941b7-505c-5ce2-91a0-b180930ec8a9","age":40,"name":"Bob"},{"_docID":"bae-3fc941b7-505c-5ce2-91a0-b180930ec8a9","_docIDNew":"bae-3fc941b7-505c-5ce2-91a0-b180930ec8a9","age":30,"name":"John"}]}`),
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

	// Verify structure instead of exact docIDs
	require.Contains(t, fileMap, "Address")
	require.NotContains(t, fileMap, "User") // Should only have Address collection after overwrite

	addresses, _ := fileMap["Address"].([]any)
	require.Len(t, addresses, 1)

	// Verify Address document
	address, _ := addresses[0].(map[string]any)
	require.Contains(t, address, "_docID")
	require.Contains(t, address, "_docIDNew")
	require.Equal(t, "Toronto", address["city"])
	require.Equal(t, "101 Maple St", address["street"])
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

	// First, create documents to get their actual docIDs
	col1, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc1, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Bob", "age": 40}`), col1.Version())
	require.NoError(t, err)
	bobID := doc1.ID().String()

	doc2, err := client.NewDocFromJSON(ctx, []byte(`{"name": "John", "age": 30}`), col1.Version())
	require.NoError(t, err)
	johnID := doc2.ID().String()

	col2, err := db.GetCollectionByName(ctx, "Address")
	require.NoError(t, err)

	doc3, err := client.NewDocFromJSON(ctx, []byte(`{"street": "101 Maple St", "city": "Toronto"}`), col2.Version())
	require.NoError(t, err)
	addressID := doc3.ID().String()

	txn, err := db.NewTxn(false)
	require.NoError(t, err)

	ctx = identity.WithContext(ctx, identity.None)
	ctx = InitContext(ctx, txn)

	filepath := t.TempDir() + "/test.json"

	// Use the actual docIDs in the import file
	importData := `{"Address":[{"_docID":"` + addressID + `","_docIDNew":"` + addressID + `","city":"Toronto","street":"101 Maple St"}],"User":[{"_docID":"` + bobID + `","_docIDNew":"` + bobID + `","age":40,"name":"Bob"},{"_docID":"` + johnID + `","_docIDNew":"` + johnID + `","age":30,"name":"John"}]}`
	err = os.WriteFile(filepath, []byte(importData), 0664)
	require.NoError(t, err)

	err = db.basicImport(ctx, filepath)
	require.NoError(t, err)
	err = txn.Commit()
	require.NoError(t, err)

	txn, err = db.NewTxn(true)
	require.NoError(t, err)

	ctx = identity.WithContext(ctx, identity.None)
	ctx = InitContext(ctx, txn)

	col1, err = db.getCollectionByName(ctx, "Address")
	require.NoError(t, err)

	key1, err := client.NewDocIDFromString(addressID)
	require.NoError(t, err)
	_, err = col1.Get(ctx, key1, false)
	require.NoError(t, err)

	col2, err = db.getCollectionByName(ctx, "User")
	require.NoError(t, err)

	key2, err := client.NewDocIDFromString(bobID)
	require.NoError(t, err)
	_, err = col2.Get(ctx, key2, false)
	require.NoError(t, err)

	key3, err := client.NewDocIDFromString(johnID)
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

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	ctx = InitContext(ctx, txn)

	filepath := t.TempDir() + "/test.json"

	err = os.WriteFile(
		filepath,
		[]byte(`["Address":[{"_docID":"bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa","_docIDNew":"bae-efd872f4-3fa4-5d0c-8a51-6339e099e9aa","city":"Toronto","street":"101 Maple St"}],"User":[{"_docID":"bae-3fc941b7-505c-5ce2-91a0-b180930ec8a9","_docIDNew":"bae-3fc941b7-505c-5ce2-91a0-b180930ec8a9","age":40,"name":"Bob"},{"_docID":"bae-3fc941b7-505c-5ce2-91a0-b180930ec8a9","_docIDNew":"bae-3fc941b7-505c-5ce2-91a0-b180930ec8a9","age":30,"name":"John"}]]`),
		0664,
	)
	require.NoError(t, err)

	err = db.basicImport(ctx, filepath)
	require.ErrorIs(t, err, ErrExpectedJSONObject)
	err = txn.Commit()
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

	txn, err := db.NewTxn(false)
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
	err = txn.Commit()
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

	txn, err := db.NewTxn(false)
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
	err = txn.Commit()
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

	txn, err := db.NewTxn(false)
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
	err = txn.Commit()
	require.NoError(t, err)
}

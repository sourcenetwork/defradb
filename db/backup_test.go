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

	"github.com/sourcenetwork/defradb/client"
)

func TestBasicExport_WithNormalFormatting_NoError(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close(ctx)

	_, err = db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}

	type Address {
		street: String
		city: String
	}`)
	require.NoError(t, err)

	doc1, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`))
	require.NoError(t, err)

	doc2, err := client.NewDocFromJSON([]byte(`{"name": "Bob", "age": 40}`))
	require.NoError(t, err)

	col1, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	err = col1.Create(ctx, doc1)
	require.NoError(t, err)

	err = col1.Create(ctx, doc2)
	require.NoError(t, err)

	doc3, err := client.NewDocFromJSON([]byte(`{"street": "101 Maple St", "city": "Toronto"}`))
	require.NoError(t, err)

	col2, err := db.GetCollectionByName(ctx, "Address")
	require.NoError(t, err)

	err = col2.Create(ctx, doc3)
	require.NoError(t, err)

	txn, err := db.NewTxn(ctx, true)
	require.NoError(t, err)

	filepath := t.TempDir() + "/test.json"
	err = db.basicExport(ctx, txn, &client.BackupConfig{Filepath: filepath})
	require.NoError(t, err)

	b, err := os.ReadFile(filepath)
	require.NoError(t, err)
	fileMap := map[string]any{}
	err = json.Unmarshal(b, &fileMap)
	require.NoError(t, err)

	expectedMap := map[string]any{}
	data := []byte(`{"Address":[{"_key":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","_newKey":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","city":"Toronto","street":"101 Maple St"}],"User":[{"_key":"bae-b94880d1-e6d2-542f-b9e0-5a369fafd0df","_newKey":"bae-b94880d1-e6d2-542f-b9e0-5a369fafd0df","age":40,"name":"Bob"},{"_key":"bae-e933420a-988a-56f8-8952-6c245aebd519","_newKey":"bae-e933420a-988a-56f8-8952-6c245aebd519","age":30,"name":"John"}]}`)
	err = json.Unmarshal(data, &expectedMap)
	require.NoError(t, err)
	require.EqualValues(t, expectedMap, fileMap)
}

func TestBasicExport_WithPrettyFormatting_NoError(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close(ctx)

	_, err = db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}

	type Address {
		street: String
		city: String
	}`)
	require.NoError(t, err)

	doc1, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`))
	require.NoError(t, err)

	doc2, err := client.NewDocFromJSON([]byte(`{"name": "Bob", "age": 40}`))
	require.NoError(t, err)

	col1, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	err = col1.Create(ctx, doc1)
	require.NoError(t, err)

	err = col1.Create(ctx, doc2)
	require.NoError(t, err)

	doc3, err := client.NewDocFromJSON([]byte(`{"street": "101 Maple St", "city": "Toronto"}`))
	require.NoError(t, err)

	col2, err := db.GetCollectionByName(ctx, "Address")
	require.NoError(t, err)

	err = col2.Create(ctx, doc3)
	require.NoError(t, err)

	txn, err := db.NewTxn(ctx, true)
	require.NoError(t, err)

	filepath := t.TempDir() + "/test.json"
	err = db.basicExport(ctx, txn, &client.BackupConfig{Filepath: filepath, Pretty: true})
	require.NoError(t, err)

	b, err := os.ReadFile(filepath)
	require.NoError(t, err)
	fileMap := map[string]any{}
	err = json.Unmarshal(b, &fileMap)
	require.NoError(t, err)

	expectedMap := map[string]any{}
	data := []byte(`{"Address":[{"_key":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","_newKey":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","city":"Toronto","street":"101 Maple St"}],"User":[{"_key":"bae-b94880d1-e6d2-542f-b9e0-5a369fafd0df","_newKey":"bae-b94880d1-e6d2-542f-b9e0-5a369fafd0df","age":40,"name":"Bob"},{"_key":"bae-e933420a-988a-56f8-8952-6c245aebd519","_newKey":"bae-e933420a-988a-56f8-8952-6c245aebd519","age":30,"name":"John"}]}`)
	err = json.Unmarshal(data, &expectedMap)
	require.NoError(t, err)
	require.EqualValues(t, expectedMap, fileMap)
}

func TestBasicExport_WithSingleCollection_NoError(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close(ctx)

	_, err = db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}

	type Address {
		street: String
		city: String
	}`)
	require.NoError(t, err)

	doc1, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`))
	require.NoError(t, err)

	doc2, err := client.NewDocFromJSON([]byte(`{"name": "Bob", "age": 40}`))
	require.NoError(t, err)

	col1, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	err = col1.Create(ctx, doc1)
	require.NoError(t, err)

	err = col1.Create(ctx, doc2)
	require.NoError(t, err)

	doc3, err := client.NewDocFromJSON([]byte(`{"street": "101 Maple St", "city": "Toronto"}`))
	require.NoError(t, err)

	col2, err := db.GetCollectionByName(ctx, "Address")
	require.NoError(t, err)

	err = col2.Create(ctx, doc3)
	require.NoError(t, err)

	txn, err := db.NewTxn(ctx, true)
	require.NoError(t, err)

	filepath := t.TempDir() + "/test.json"
	err = db.basicExport(ctx, txn, &client.BackupConfig{Filepath: filepath, Collections: []string{"Address"}})
	require.NoError(t, err)

	b, err := os.ReadFile(filepath)
	require.NoError(t, err)
	fileMap := map[string]any{}
	err = json.Unmarshal(b, &fileMap)
	require.NoError(t, err)

	expectedMap := map[string]any{}
	data := []byte(`{"Address":[{"_key":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","_newKey":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","city":"Toronto","street":"101 Maple St"}]}`)
	err = json.Unmarshal(data, &expectedMap)
	require.NoError(t, err)
	require.EqualValues(t, expectedMap, fileMap)
}

func TestBasicImport_WithMultipleCollectionsAndObjects_NoError(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close(ctx)

	_, err = db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}

	type Address {
		street: String
		city: String
	}`)
	require.NoError(t, err)

	txn, err := db.NewTxn(ctx, true)
	require.NoError(t, err)

	filepath := t.TempDir() + "/test.json"

	err = os.WriteFile(
		filepath,
		[]byte(`{"Address":[{"_key":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","_newKey":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","city":"Toronto","street":"101 Maple St"}],"User":[{"_key":"bae-b94880d1-e6d2-542f-b9e0-5a369fafd0df","_newKey":"bae-b94880d1-e6d2-542f-b9e0-5a369fafd0df","age":40,"name":"Bob"},{"_key":"bae-e933420a-988a-56f8-8952-6c245aebd519","_newKey":"bae-e933420a-988a-56f8-8952-6c245aebd519","age":30,"name":"John"}]}`),
		0664,
	)
	require.NoError(t, err)

	err = db.basicImport(ctx, txn, filepath)
	require.NoError(t, err)

	col1, err := db.getCollectionByName(ctx, txn, "Address")
	require.NoError(t, err)

	key1, err := client.NewDocKeyFromString("bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f")
	require.NoError(t, err)
	_, err = col1.Get(ctx, key1, false)
	require.NoError(t, err)

	col2, err := db.getCollectionByName(ctx, txn, "User")
	require.NoError(t, err)

	key2, err := client.NewDocKeyFromString("bae-b94880d1-e6d2-542f-b9e0-5a369fafd0df")
	require.NoError(t, err)
	_, err = col2.Get(ctx, key2, false)
	require.NoError(t, err)

	key3, err := client.NewDocKeyFromString("bae-e933420a-988a-56f8-8952-6c245aebd519")
	require.NoError(t, err)
	_, err = col2.Get(ctx, key3, false)
	require.NoError(t, err)
}

func TestBasicImport_WithJSONArray_ReturnError(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close(ctx)

	_, err = db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}

	type Address {
		street: String
		city: String
	}`)
	require.NoError(t, err)

	txn, err := db.NewTxn(ctx, true)
	require.NoError(t, err)

	filepath := t.TempDir() + "/test.json"

	err = os.WriteFile(
		filepath,
		[]byte(`["Address":[{"_key":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","_newKey":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","city":"Toronto","street":"101 Maple St"}],"User":[{"_key":"bae-b94880d1-e6d2-542f-b9e0-5a369fafd0df","_newKey":"bae-b94880d1-e6d2-542f-b9e0-5a369fafd0df","age":40,"name":"Bob"},{"_key":"bae-e933420a-988a-56f8-8952-6c245aebd519","_newKey":"bae-e933420a-988a-56f8-8952-6c245aebd519","age":30,"name":"John"}]]`),
		0664,
	)
	require.NoError(t, err)

	err = db.basicImport(ctx, txn, filepath)
	require.ErrorIs(t, err, ErrExpectedJSONObject)
}

func TestBasicImport_WithObjectCollection_ReturnError(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close(ctx)

	_, err = db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}

	type Address {
		street: String
		city: String
	}`)
	require.NoError(t, err)

	txn, err := db.NewTxn(ctx, true)
	require.NoError(t, err)

	filepath := t.TempDir() + "/test.json"

	err = os.WriteFile(
		filepath,
		[]byte(`{"Address":{"_key":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","_newKey":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","city":"Toronto","street":"101 Maple St"}}`),
		0664,
	)
	require.NoError(t, err)

	err = db.basicImport(ctx, txn, filepath)
	require.ErrorIs(t, err, ErrExpectedJSONArray)
}

func TestBasicImport_WithInvalidFilepath_ReturnError(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close(ctx)

	_, err = db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}

	type Address {
		street: String
		city: String
	}`)
	require.NoError(t, err)

	txn, err := db.NewTxn(ctx, true)
	require.NoError(t, err)

	filepath := t.TempDir() + "/test.json"

	err = os.WriteFile(
		filepath,
		[]byte(`{"Address":{"_key":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","_newKey":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","city":"Toronto","street":"101 Maple St"}}`),
		0664,
	)
	require.NoError(t, err)

	wrongFilepath := t.TempDir() + "/some/test.json"
	err = db.basicImport(ctx, txn, wrongFilepath)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestBasicImport_WithInvalidCollection_ReturnError(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close(ctx)

	_, err = db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}

	type Address {
		street: String
		city: String
	}`)
	require.NoError(t, err)

	txn, err := db.NewTxn(ctx, true)
	require.NoError(t, err)

	filepath := t.TempDir() + "/test.json"

	err = os.WriteFile(
		filepath,
		[]byte(`{"Addresses":{"_key":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","_newKey":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","city":"Toronto","street":"101 Maple St"}}`),
		0664,
	)
	require.NoError(t, err)

	err = db.basicImport(ctx, txn, filepath)
	require.ErrorIs(t, err, ErrFailedToGetCollection)
}

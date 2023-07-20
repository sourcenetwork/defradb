// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

func TestDBExportCmd_WithNoArgument_ReturnError(t *testing.T) {
	cfg := getTestConfig(t)

	dbExportCmd := MakeDBExportCommand(cfg)
	err := dbExportCmd.ValidateArgs([]string{})
	require.ErrorIs(t, err, ErrInvalidArgumentLength)
}

func TestDBExportCmd_WithInvalidExportFormat_ReturnError(t *testing.T) {
	cfg := getTestConfig(t)
	dbExportCmd := MakeDBExportCommand(cfg)

	filepath := t.TempDir() + "/test.json"

	dbExportCmd.Flags().Set("format", "invalid")
	err := dbExportCmd.RunE(dbExportCmd, []string{filepath})
	require.ErrorIs(t, err, ErrInvalidExportFormat)
}

func TestDBExportCmd_IfInvalidAddress_ReturnError(t *testing.T) {
	cfg := getTestConfig(t)
	cfg.API.Address = "invalid address"

	filepath := t.TempDir() + "/test.json"

	dbExportCmd := MakeDBExportCommand(cfg)
	err := dbExportCmd.RunE(dbExportCmd, []string{filepath})
	require.ErrorIs(t, err, NewErrFailedToJoinEndpoint(err))
}

func TestDBExportCmd_WithEmptyDatastore_NoError(t *testing.T) {
	cfg, _, close := startTestNode(t)
	defer close()

	filepath := t.TempDir() + "/test.json"

	outputBuf, revertOutput := simulateConsoleOutput(t)
	defer revertOutput()

	dbExportCmd := MakeDBExportCommand(cfg)
	err := dbExportCmd.RunE(dbExportCmd, []string{filepath})
	require.NoError(t, err)

	logLines, err := parseLines(outputBuf)
	require.NoError(t, err)
	require.True(t, lineHas(logLines, "msg", "Data exported for all collections"))

	b, err := os.ReadFile(filepath)
	require.NoError(t, err)

	require.Len(t, b, 2) // file should be an empty json object
}

func TestDBExportCmd_WithInvalidCollection_ReturnError(t *testing.T) {
	cfg, _, close := startTestNode(t)
	defer close()

	filepath := t.TempDir() + "/test.json"

	outputBuf, revertOutput := simulateConsoleOutput(t)
	defer revertOutput()

	dbExportCmd := MakeDBExportCommand(cfg)
	dbExportCmd.Flags().Set("collections", "User")
	err := dbExportCmd.RunE(dbExportCmd, []string{filepath})
	require.NoError(t, err)

	logLines, err := parseLines(outputBuf)
	require.NoError(t, err)
	require.True(t, lineHas(logLines, "msg", "Failed to export data"))
}

func TestDBExportCmd_WithAllCollection_NoError(t *testing.T) {
	ctx := context.Background()

	cfg, di, close := startTestNode(t)
	defer close()

	_, err := di.db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`))
	require.NoError(t, err)

	col, err := di.db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	err = col.Create(ctx, doc)
	require.NoError(t, err)

	filepath := t.TempDir() + "/test.json"

	outputBuf, revertOutput := simulateConsoleOutput(t)
	defer revertOutput()

	dbExportCmd := MakeDBExportCommand(cfg)
	err = dbExportCmd.RunE(dbExportCmd, []string{filepath})
	require.NoError(t, err)

	logLines, err := parseLines(outputBuf)
	require.NoError(t, err)
	require.True(t, lineHas(logLines, "msg", "Data exported for all collections"))

	b, err := os.ReadFile(filepath)
	require.NoError(t, err)

	require.Equal(
		t,
		`{"User":[{"_key":"bae-e933420a-988a-56f8-8952-6c245aebd519","_newKey":"bae-e933420a-988a-56f8-8952-6c245aebd519","age":30,"name":"John"}]}`,
		string(b),
	)
}

func TestDBExportCmd_WithAllCollectionAndPrettyFormating_NoError(t *testing.T) {
	ctx := context.Background()

	cfg, di, close := startTestNode(t)
	defer close()

	_, err := di.db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`))
	require.NoError(t, err)

	col, err := di.db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	err = col.Create(ctx, doc)
	require.NoError(t, err)

	filepath := t.TempDir() + "/test.json"

	outputBuf, revertOutput := simulateConsoleOutput(t)
	defer revertOutput()

	dbExportCmd := MakeDBExportCommand(cfg)
	dbExportCmd.Flags().Set("pretty", "true")
	err = dbExportCmd.RunE(dbExportCmd, []string{filepath})
	require.NoError(t, err)

	logLines, err := parseLines(outputBuf)
	require.NoError(t, err)
	require.True(t, lineHas(logLines, "msg", "Data exported for all collections"))

	b, err := os.ReadFile(filepath)
	require.NoError(t, err)

	require.Equal(
		t,
		`{
  "User": [
    {
      "_key": "bae-e933420a-988a-56f8-8952-6c245aebd519",
      "_newKey": "bae-e933420a-988a-56f8-8952-6c245aebd519",
      "age": 30,
      "name": "John"
    }
  ]
}`,
		string(b),
	)
}

func TestDBExportCmd_WithSingleCollection_NoError(t *testing.T) {
	ctx := context.Background()

	cfg, di, close := startTestNode(t)
	defer close()

	_, err := di.db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`))
	require.NoError(t, err)

	col, err := di.db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	err = col.Create(ctx, doc)
	require.NoError(t, err)

	filepath := t.TempDir() + "/test.json"

	outputBuf, revertOutput := simulateConsoleOutput(t)
	defer revertOutput()

	dbExportCmd := MakeDBExportCommand(cfg)
	dbExportCmd.Flags().Set("collections", "User")
	err = dbExportCmd.RunE(dbExportCmd, []string{filepath})
	require.NoError(t, err)

	logLines, err := parseLines(outputBuf)
	require.NoError(t, err)
	require.True(t, lineHas(logLines, "msg", "Data exported for collection User"))

	b, err := os.ReadFile(filepath)
	require.NoError(t, err)

	require.Equal(
		t,
		`{"User":[{"_key":"bae-e933420a-988a-56f8-8952-6c245aebd519","_newKey":"bae-e933420a-988a-56f8-8952-6c245aebd519","age":30,"name":"John"}]}`,
		string(b),
	)
}

func TestDBExportCmd_WithMultipleCollections_NoError(t *testing.T) {
	ctx := context.Background()

	cfg, di, close := startTestNode(t)
	defer close()

	_, err := di.db.AddSchema(ctx, `type User {
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

	col1, err := di.db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	err = col1.Create(ctx, doc1)
	require.NoError(t, err)

	doc2, err := client.NewDocFromJSON([]byte(`{"street": "101 Maple St", "city": "Toronto"}`))
	require.NoError(t, err)

	col2, err := di.db.GetCollectionByName(ctx, "Address")
	require.NoError(t, err)

	err = col2.Create(ctx, doc2)
	require.NoError(t, err)

	filepath := t.TempDir() + "/test.json"

	outputBuf, revertOutput := simulateConsoleOutput(t)
	defer revertOutput()

	dbExportCmd := MakeDBExportCommand(cfg)
	dbExportCmd.Flags().Set("collections", "User, Address")
	err = dbExportCmd.RunE(dbExportCmd, []string{filepath})
	require.NoError(t, err)

	logLines, err := parseLines(outputBuf)
	require.NoError(t, err)
	require.True(t, lineHas(logLines, "msg", "Data exported for collections User, Address"))

	b, err := os.ReadFile(filepath)
	require.NoError(t, err)
	fileMap := map[string]any{}
	err = json.Unmarshal(b, &fileMap)
	require.NoError(t, err)

	expectedMap := map[string]any{}
	data := []byte(`{"Address":[{"_key":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","_newKey":"bae-8096f2c1-ea4c-5226-8ba5-17fc4b68ac1f","city":"Toronto","street":"101 Maple St"}],"User":[{"_key":"bae-e933420a-988a-56f8-8952-6c245aebd519","_newKey":"bae-e933420a-988a-56f8-8952-6c245aebd519","age":30,"name":"John"}]}`)
	err = json.Unmarshal(data, &expectedMap)
	require.NoError(t, err)

	require.EqualValues(t, expectedMap, fileMap)
}

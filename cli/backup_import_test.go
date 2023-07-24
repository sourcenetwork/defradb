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
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

func TestBackupImportCmd_WithNoArgument_ReturnError(t *testing.T) {
	cfg := getTestConfig(t)
	setTestingAddresses(cfg)

	dbImportCmd := MakeBackupImportCommand(cfg)
	err := dbImportCmd.ValidateArgs([]string{})
	require.ErrorIs(t, err, ErrInvalidArgumentLength)
}

func TestBackupImportCmd_IfInvalidAddress_ReturnError(t *testing.T) {
	cfg := getTestConfig(t)
	cfg.API.Address = "invalid address"

	filepath := t.TempDir() + "/test.json"

	dbImportCmd := MakeBackupImportCommand(cfg)
	err := dbImportCmd.RunE(dbImportCmd, []string{filepath})
	require.ErrorIs(t, err, NewErrFailedToJoinEndpoint(err))
}

func TestBackupImportCmd_WithNonExistantFile_ReturnError(t *testing.T) {
	cfg, _, close := startTestNode(t)
	defer close()

	filepath := t.TempDir() + "/test.json"

	outputBuf, revertOutput := simulateConsoleOutput(t)
	defer revertOutput()

	dbImportCmd := MakeBackupImportCommand(cfg)
	err := dbImportCmd.RunE(dbImportCmd, []string{filepath})
	require.NoError(t, err)

	logLines, err := parseLines(outputBuf)
	require.NoError(t, err)
	require.True(t, lineHas(logLines, "msg", "Failed to import data"))
}

func TestBackupImportCmd_WithEmptyDatastore_ReturnError(t *testing.T) {
	cfg, _, close := startTestNode(t)
	defer close()

	filepath := t.TempDir() + "/test.json"

	err := os.WriteFile(
		filepath,
		[]byte(`{"User":[{"_key":"bae-e933420a-988a-56f8-8952-6c245aebd519","_newKey":"bae-e933420a-988a-56f8-8952-6c245aebd519","age":30,"name":"John"}]}`),
		0664,
	)
	require.NoError(t, err)

	outputBuf, revertOutput := simulateConsoleOutput(t)
	defer revertOutput()

	dbImportCmd := MakeBackupImportCommand(cfg)
	err = dbImportCmd.RunE(dbImportCmd, []string{filepath})
	require.NoError(t, err)

	logLines, err := parseLines(outputBuf)
	require.NoError(t, err)
	require.True(t, lineHas(logLines, "msg", "Failed to import data"))
}

func TestBackupImportCmd_WithExistingCollection_NoError(t *testing.T) {
	ctx := context.Background()

	cfg, di, close := startTestNode(t)
	defer close()

	_, err := di.db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	filepath := t.TempDir() + "/test.json"

	err = os.WriteFile(
		filepath,
		[]byte(`{"User":[{"_key":"bae-e933420a-988a-56f8-8952-6c245aebd519","_newKey":"bae-e933420a-988a-56f8-8952-6c245aebd519","age":30,"name":"John"}]}`),
		0664,
	)
	require.NoError(t, err)

	outputBuf, revertOutput := simulateConsoleOutput(t)
	defer revertOutput()

	dbImportCmd := MakeBackupImportCommand(cfg)
	err = dbImportCmd.RunE(dbImportCmd, []string{filepath})
	require.NoError(t, err)

	logLines, err := parseLines(outputBuf)
	require.NoError(t, err)
	require.True(t, lineHas(logLines, "msg", "Successfully imported data from file"))

	col, err := di.db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	key, err := client.NewDocKeyFromString("bae-e933420a-988a-56f8-8952-6c245aebd519")
	require.NoError(t, err)
	doc, err := col.Get(ctx, key, false)
	require.NoError(t, err)

	val, err := doc.Get("name")
	require.NoError(t, err)

	require.Equal(t, "John", val.(string))
}

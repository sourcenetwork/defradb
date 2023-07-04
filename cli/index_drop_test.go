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
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndexDropCmd_IfInvalidAddress_ReturnError(t *testing.T) {
	cfg := getTestConfig(t)
	cfg.API.Address = "invalid address"
	indexDropCmd := MakeIndexDropCommand(cfg)

	indexDropCmd.SetArgs([]string{"--collection", "User", "--name", "users_name_index"})
	err := indexDropCmd.Execute()
	require.ErrorIs(t, err, NewErrFailedToJoinEndpoint(err))
}

func TestIndexDropCmd_IfNoCollection_ReturnError(t *testing.T) {
	cfg, close := startNode(t)
	defer close()
	indexDropCmd := MakeIndexDropCommand(cfg)

	outputBuf := bytes.NewBufferString("")
	indexDropCmd.SetOut(outputBuf)

	indexDropCmd.SetArgs([]string{"--collection", "User", "--name", "users_name_index"})
	err := indexDropCmd.Execute()
	require.NoError(t, err)

	out, err := io.ReadAll(outputBuf)
	require.NoError(t, err)

	r := make(map[string]any)
	err = json.Unmarshal(out, &r)
	require.NoError(t, err)

	_, hasErrors := r["errors"]
	assert.True(t, hasErrors, "command should return error")
}

func TestIndexDropCmd_IfNoErrors_ShouldReturnData(t *testing.T) {
	cfg, close := startNode(t)
	defer close()

	execAddSchemaCmd(t, cfg, `type User { name: String }`)
	execCreateIndexCmd(t, cfg, "User", "name", "users_name_index")

	indexDropCmd := MakeIndexDropCommand(cfg)
	outputBuf := bytes.NewBufferString("")
	indexDropCmd.SetOut(outputBuf)

	indexDropCmd.SetArgs([]string{"--collection", "User", "--name", "users_name_index"})
	err := indexDropCmd.Execute()
	require.NoError(t, err)

	out, err := io.ReadAll(outputBuf)
	require.NoError(t, err)

	r := make(map[string]any)
	err = json.Unmarshal(out, &r)
	require.NoError(t, err)

	_, hasData := r["data"]
	assert.True(t, hasData, "command should return data")
}

func TestIndexDropCmd_WithConsoleOutputIfNoCollection_ReturnError(t *testing.T) {
	cfg, close := startNode(t)
	defer close()
	indexDropCmd := MakeIndexDropCommand(cfg)

	outputBuf, revertOutput := simulateConsoleOutput(t)
	defer revertOutput()

	indexDropCmd.SetArgs([]string{"--collection", "User", "--name", "users_name_index"})
	err := indexDropCmd.Execute()
	require.NoError(t, err)

	logLines, err := parseLines(outputBuf)
	require.NoError(t, err)
	assert.True(t, hasLogWithKey(logLines, "Errors"))
}

func TestIndexDropCmd_WithConsoleOutputIfNoErrors_ShouldReturnData(t *testing.T) {
	cfg, close := startNode(t)
	defer close()

	execAddSchemaCmd(t, cfg, `type User { name: String }`)
	execCreateIndexCmd(t, cfg, "User", "name", "users_name_index")

	indexDropCmd := MakeIndexDropCommand(cfg)
	indexDropCmd.SetArgs([]string{"--collection", "User", "--name", "users_name_index"})

	outputBuf, revertOutput := simulateConsoleOutput(t)
	defer revertOutput()

	err := indexDropCmd.Execute()
	require.NoError(t, err)

	logLines, err := parseLines(outputBuf)
	require.NoError(t, err)
	require.Len(t, logLines, 1)
	assert.Equal(t, "success", logLines[0]["Result"])

	assert.False(t, hasLogWithKey(logLines, "Errors"))
}

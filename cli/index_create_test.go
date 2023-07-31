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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/logging"
)

const randomMultiaddr = "/ip4/0.0.0.0/tcp/0"

func getTestConfig(t *testing.T) *config.Config {
	cfg := config.DefaultConfig()
	dir := t.TempDir()
	cfg.Datastore.Store = "memory"
	cfg.Datastore.Badger.Path = dir
	cfg.Net.P2PDisabled = false
	cfg.Net.P2PAddress = randomMultiaddr
	cfg.Net.RPCAddress = "0.0.0.0:0"
	cfg.Net.TCPAddress = randomMultiaddr
	return cfg
}

func startTestNode(t *testing.T) (*config.Config, *defraInstance, func()) {
	cfg := getTestConfig(t)
	setTestingAddresses(cfg)

	ctx := context.Background()
	di, err := start(ctx, cfg)
	require.NoError(t, err)
	return cfg, di, func() { di.close(ctx) }
}

func parseLines(r io.Reader) ([]map[string]any, error) {
	fileScanner := bufio.NewScanner(r)

	fileScanner.Split(bufio.ScanLines)

	logLines := []map[string]any{}
	for fileScanner.Scan() {
		loggedLine := make(map[string]any)
		err := json.Unmarshal(fileScanner.Bytes(), &loggedLine)
		if err != nil {
			return nil, err
		}
		logLines = append(logLines, loggedLine)
	}

	return logLines, nil
}

func lineHas(lines []map[string]any, key, value string) bool {
	for _, line := range lines {
		if line[key] == value {
			return true
		}
	}
	return false
}

func simulateConsoleOutput(t *testing.T) (*bytes.Buffer, func()) {
	b := &bytes.Buffer{}
	log.ApplyConfig(logging.Config{
		EncoderFormat: logging.NewEncoderFormatOption(logging.JSON),
		Pipe:          b,
	})

	f, err := os.CreateTemp(t.TempDir(), "tmpFile")
	require.NoError(t, err)
	originalStdout := os.Stdout
	os.Stdout = f

	return b, func() {
		os.Stdout = originalStdout
		f.Close()
		os.Remove(f.Name())
	}
}

func execAddSchemaCmd(t *testing.T, cfg *config.Config, schema string) {
	addSchemaCmd := MakeSchemaAddCommand(cfg)
	err := addSchemaCmd.RunE(addSchemaCmd, []string{schema})
	require.NoError(t, err)
}

func execCreateIndexCmd(t *testing.T, cfg *config.Config, collection, fields, name string) {
	indexCreateCmd := MakeIndexCreateCommand(cfg)
	indexCreateCmd.SetArgs([]string{
		"--collection", collection,
		"--fields", fields,
		"--name", name,
	})
	err := indexCreateCmd.Execute()
	require.NoError(t, err)
}

func hasLogWithKey(logLines []map[string]any, key string) bool {
	for _, logLine := range logLines {
		if _, ok := logLine[key]; ok {
			return true
		}
	}
	return false
}

func TestIndexCreateCmd_IfInvalidAddress_ReturnError(t *testing.T) {
	cfg := getTestConfig(t)
	cfg.API.Address = "invalid address"
	indexCreateCmd := MakeIndexCreateCommand(cfg)

	indexCreateCmd.SetArgs([]string{
		"--collection", "User",
		"--fields", "Name",
		"--name", "users_name_index",
	})
	err := indexCreateCmd.Execute()
	require.ErrorIs(t, err, NewErrFailedToJoinEndpoint(err))
}

func TestIndexCreateCmd_IfNoCollection_ReturnError(t *testing.T) {
	cfg, _, close := startTestNode(t)
	defer close()
	indexCreateCmd := MakeIndexCreateCommand(cfg)

	outputBuf := bytes.NewBufferString("")
	indexCreateCmd.SetOut(outputBuf)

	indexCreateCmd.SetArgs([]string{
		"--collection", "User",
		"--fields", "Name",
		"--name", "users_name_index",
	})
	err := indexCreateCmd.Execute()
	require.NoError(t, err)

	out, err := io.ReadAll(outputBuf)
	require.NoError(t, err)

	r := make(map[string]any)
	err = json.Unmarshal(out, &r)
	require.NoError(t, err)

	_, hasErrors := r["errors"]
	assert.True(t, hasErrors, "command should return error")
}

func TestIndexCreateCmd_IfNoErrors_ReturnData(t *testing.T) {
	cfg, _, close := startTestNode(t)
	defer close()

	execAddSchemaCmd(t, cfg, `type User { name: String }`)

	indexCreateCmd := MakeIndexCreateCommand(cfg)
	outputBuf := bytes.NewBufferString("")
	indexCreateCmd.SetOut(outputBuf)

	indexCreateCmd.SetArgs([]string{
		"--collection", "User",
		"--fields", "name",
		"--name", "users_name_index",
	})
	err := indexCreateCmd.Execute()
	require.NoError(t, err)

	out, err := io.ReadAll(outputBuf)
	require.NoError(t, err)

	r := make(map[string]any)
	err = json.Unmarshal(out, &r)
	require.NoError(t, err)

	_, hasData := r["data"]
	assert.True(t, hasData, "command should return data")
}

func TestIndexCreateCmd_WithConsoleOutputIfNoCollection_ReturnError(t *testing.T) {
	cfg, _, close := startTestNode(t)
	defer close()
	indexCreateCmd := MakeIndexCreateCommand(cfg)
	indexCreateCmd.SetArgs([]string{
		"--collection", "User",
		"--fields", "Name",
		"--name", "users_name_index",
	})

	outputBuf, revertOutput := simulateConsoleOutput(t)
	defer revertOutput()

	err := indexCreateCmd.Execute()
	require.NoError(t, err)

	logLines, err := parseLines(outputBuf)
	require.NoError(t, err)
	assert.True(t, hasLogWithKey(logLines, "Errors"))
}

func TestIndexCreateCmd_WithConsoleOutputIfNoErrors_ReturnData(t *testing.T) {
	cfg, _, close := startTestNode(t)
	defer close()

	execAddSchemaCmd(t, cfg, `type User { name: String }`)

	const indexName = "users_name_index"
	indexCreateCmd := MakeIndexCreateCommand(cfg)
	indexCreateCmd.SetArgs([]string{
		"--collection", "User",
		"--fields", "name",
		"--name", indexName,
	})

	outputBuf, revertOutput := simulateConsoleOutput(t)
	defer revertOutput()

	err := indexCreateCmd.Execute()
	require.NoError(t, err)

	logLines, err := parseLines(outputBuf)
	require.NoError(t, err)
	require.Len(t, logLines, 1)
	result, ok := logLines[0]["Index"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, indexName, result["Name"])

	assert.False(t, hasLogWithKey(logLines, "Errors"))
}

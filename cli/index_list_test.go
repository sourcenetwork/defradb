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

func TestIndexListCmd_IfInvalidAddress_ReturnError(t *testing.T) {
	cfg := getTestConfig(t)
	cfg.API.Address = "invalid address"
	indexCreateCmd := MakeIndexListCommand(cfg)

	err := indexCreateCmd.RunE(indexCreateCmd, nil)
	require.ErrorIs(t, err, NewErrFailedToJoinEndpoint(err))
}

func TestIndexListCmd_IfNoErrors_ShouldReturnData(t *testing.T) {
	cfg, close := startNode(t)
	defer close()

	execAddSchemaCmd(t, cfg, `type User { name: String }`)
	execCreateIndexCmd(t, cfg, "User", "name", "users_name_index")

	indexListCmd := MakeIndexListCommand(cfg)
	outputBuf := bytes.NewBufferString("")
	indexListCmd.SetOut(outputBuf)

	err := indexListCmd.Execute()
	require.NoError(t, err)

	out, err := io.ReadAll(outputBuf)
	require.NoError(t, err)

	r := make(map[string]any)
	err = json.Unmarshal(out, &r)
	require.NoError(t, err)

	_, hasData := r["data"]
	assert.True(t, hasData, "command should return data")
}

func TestIndexListCmd_WithConsoleOutputIfCollectionDoesNotExist_ReturnError(t *testing.T) {
	cfg, close := startNode(t)
	defer close()

	indexListCmd := MakeIndexListCommand(cfg)
	indexListCmd.SetArgs([]string{"--collection", "User"})

	outputBuf, revertOutput := simulateConsoleOutput(t)
	defer revertOutput()

	err := indexListCmd.Execute()
	require.NoError(t, err)

	logLines, err := parseLines(outputBuf)
	require.NoError(t, err)
	require.True(t, hasLogWithKey(logLines, "Errors"))
}

func TestIndexListCmd_WithConsoleOutputIfCollectionIsGiven_ReturnCollectionList(t *testing.T) {
	cfg, close := startNode(t)
	defer close()

	const indexName = "users_name_index"
	execAddSchemaCmd(t, cfg, `type User { name: String }`)
	execCreateIndexCmd(t, cfg, "User", "name", indexName)

	indexListCmd := MakeIndexListCommand(cfg)
	indexListCmd.SetArgs([]string{"--collection", "User"})

	outputBuf, revertOutput := simulateConsoleOutput(t)
	defer revertOutput()

	err := indexListCmd.Execute()
	require.NoError(t, err)

	logLines, err := parseLines(outputBuf)
	require.NoError(t, err)
	require.Len(t, logLines, 1)
	resultList, ok := logLines[0]["Indexes"].([]any)
	require.True(t, ok)
	require.Len(t, resultList, 1)
	result, ok := resultList[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, indexName, result["Name"])

	assert.False(t, hasLogWithKey(logLines, "Errors"))
}

func TestIndexListCmd_WithConsoleOutputIfNoArgs_ReturnAllIndexes(t *testing.T) {
	cfg, close := startNode(t)
	defer close()

	const userIndexName = "users_name_index"
	const productIndexName = "product_price_index"
	execAddSchemaCmd(t, cfg, `type User { name: String }`)
	execAddSchemaCmd(t, cfg, `type Product { price: Int }`)
	execCreateIndexCmd(t, cfg, "User", "name", userIndexName)
	execCreateIndexCmd(t, cfg, "Product", "price", productIndexName)

	indexListCmd := MakeIndexListCommand(cfg)

	outputBuf, revertOutput := simulateConsoleOutput(t)
	defer revertOutput()

	err := indexListCmd.Execute()
	require.NoError(t, err)

	logLines, err := parseLines(outputBuf)
	require.NoError(t, err)
	require.Len(t, logLines, 1)
	resultCollections, ok := logLines[0]["Collections"].(map[string]any)
	require.True(t, ok)

	userCollection, ok := resultCollections["User"].([]any)
	require.True(t, ok)
	require.Len(t, userCollection, 1)
	userIndex, ok := userCollection[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, userIndexName, userIndex["Name"])

	productCollection, ok := resultCollections["Product"].([]any)
	require.True(t, ok)
	require.Len(t, productCollection, 1)
	productIndex, ok := productCollection[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, productIndexName, productIndex["Name"])

	assert.False(t, hasLogWithKey(logLines, "Errors"))
}

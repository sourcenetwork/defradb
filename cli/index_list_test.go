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

	b := bytes.NewBufferString("")
	indexCreateCmd.SetOut(b)

	err := indexCreateCmd.RunE(indexCreateCmd, nil)
	require.ErrorIs(t, err, NewErrFailedToJoinEndpoint(err))
}

func TestIndexListCmd_IfNonExistingAddress_ReturnError(t *testing.T) {
	cfg := getTestConfig(t)
	cfg.API.Address = "none"
	indexCreateCmd := MakeIndexListCommand(cfg)

	b := bytes.NewBufferString("")
	indexCreateCmd.SetOut(b)

	err := indexCreateCmd.RunE(indexCreateCmd, nil)
	require.ErrorIs(t, err, NewErrFailedToSendRequest(err))
}

func TestIndexListCmd_IfNoErrors_ShouldReturnData(t *testing.T) {
	cfg, close := startNode(t)
	defer close()

	addSchemaCmd := MakeSchemaAddCommand(cfg)
	err := addSchemaCmd.RunE(addSchemaCmd, []string{`type User { name: String }`})
	if err != nil {
		t.Fatal(err)
	}

	indexCreateCmd := MakeIndexCreateCommand(cfg)
	indexCreateCmd.SetArgs([]string{"--collection", "User",
		"--fields", "Name", "--name", "users_name_index"})
	err = indexCreateCmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	indexListCmd := MakeIndexListCommand(cfg)
	b := bytes.NewBufferString("")
	indexListCmd.SetOut(b)

	err = indexListCmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	out, err := io.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}

	r := make(map[string]any)
	err = json.Unmarshal(out, &r)
	if err != nil {
		t.Fatal(err)
	}

	_, hasData := r["data"]
	assert.True(t, hasData, "command should return data")
}

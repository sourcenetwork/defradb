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
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/config"
)

func getTestConfig(t *testing.T) *config.Config {
	cfg := config.DefaultConfig()
	dir := t.TempDir()
	cfg.Datastore.Store = "memory"
	cfg.Datastore.Badger.Path = dir
	cfg.Net.P2PDisabled = false
	return cfg
}

func startNode(t *testing.T) (*config.Config, func()) {
	cfg := getTestConfig(t)
	setTestingAddresses(cfg)

	ctx := context.Background()
	di, err := start(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	return cfg, func() { di.close(ctx) }
}

func TestIndexCreateCmd_IfInvalidAddress_ReturnError(t *testing.T) {
	cfg := getTestConfig(t)
	cfg.API.Address = "invalid address"
	indexCreateCmd := MakeIndexCreateCommand(cfg)

	b := bytes.NewBufferString("")
	indexCreateCmd.SetOut(b)

	indexCreateCmd.SetArgs([]string{"--collection", "User",
		"--fields", "Name", "--name", "users_name_index"})
	err := indexCreateCmd.Execute()
	require.ErrorIs(t, err, NewErrFailedToJoinEndpoint(err))
}

func TestIndexCreateCmd_IfNonExistingAddress_ReturnError(t *testing.T) {
	cfg := getTestConfig(t)
	cfg.API.Address = "none"
	indexCreateCmd := MakeIndexCreateCommand(cfg)

	b := bytes.NewBufferString("")
	indexCreateCmd.SetOut(b)

	indexCreateCmd.SetArgs([]string{"--collection", "User",
		"--fields", "Name", "--name", "users_name_index"})
	err := indexCreateCmd.Execute()
	require.ErrorIs(t, err, NewErrFailedToSendRequest(err))
}

func TestIndexCreateCmd_IfNoCollection_ReturnError(t *testing.T) {
	cfg, close := startNode(t)
	defer close()
	indexCreateCmd := MakeIndexCreateCommand(cfg)

	b := bytes.NewBufferString("")
	indexCreateCmd.SetOut(b)

	indexCreateCmd.SetArgs([]string{"--collection", "User",
		"--fields", "Name", "--name", "users_name_index"})
	err := indexCreateCmd.Execute()
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

	_, hasErrors := r["errors"]
	assert.True(t, hasErrors, "command should return error")
}

func TestIndexCreateCmd_IfNoErrors_ShouldReturnData(t *testing.T) {
	cfg, close := startNode(t)
	defer close()

	addSchemaCmd := MakeSchemaAddCommand(cfg)
	err := addSchemaCmd.RunE(addSchemaCmd, []string{`type User { name: String }`})
	if err != nil {
		t.Fatal(err)
	}

	indexCreateCmd := MakeIndexCreateCommand(cfg)
	b := bytes.NewBufferString("")
	indexCreateCmd.SetOut(b)

	indexCreateCmd.SetArgs([]string{"--collection", "User",
		"--fields", "name", "--name", "users_name_index"})
	err = indexCreateCmd.Execute()
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

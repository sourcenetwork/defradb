// Copyright 2022 Democratized Data Foundation
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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	httpapi "github.com/sourcenetwork/defradb/api/http"
	"github.com/sourcenetwork/defradb/config"
)

// setTestingAddresses overrides the config addresses to be the ones reserved for testing.
// Used to ensure the tests don't fail due to address clashes with the running server (with default config).
func setTestingAddresses(cfg *config.Config) {
	cfg.API.Address = "localhost:9182"
	cfg.Net.P2PAddress = "/ip4/0.0.0.0/tcp/9172"
	cfg.Net.TCPAddress = "/ip4/0.0.0.0/tcp/9162"
	cfg.Net.RPCAddress = "0.0.0.0:9162"
}

func TestGetPeerIDCmd(t *testing.T) {
	cfg := config.DefaultConfig()
	peerIDCmd := MakePeerIDCommand(cfg)
	dir := t.TempDir()
	ctx := context.Background()
	cfg.Datastore.Store = "memory"
	cfg.Datastore.Badger.Path = dir
	cfg.Net.P2PDisabled = false
	setTestingAddresses(cfg)

	di, err := start(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer di.close(ctx)

	b := bytes.NewBufferString("")
	peerIDCmd.SetOut(b)

	err = peerIDCmd.RunE(peerIDCmd, nil)
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

	assert.Equal(t, di.node.PeerID().String(), r["peerID"])
}

func TestGetPeerIDCmdWithNoP2P(t *testing.T) {
	cfg := config.DefaultConfig()
	peerIDCmd := MakePeerIDCommand(cfg)
	dir := t.TempDir()
	ctx := context.Background()
	cfg.Datastore.Store = "memory"
	cfg.Datastore.Badger.Path = dir
	cfg.Net.P2PDisabled = true
	setTestingAddresses(cfg)

	di, err := start(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer di.close(ctx)

	b := bytes.NewBufferString("")
	peerIDCmd.SetOut(b)

	err = peerIDCmd.RunE(peerIDCmd, nil)
	if err != nil {
		t.Fatal(err)
	}

	out, err := io.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}

	r := httpapi.ErrorItem{}
	err = json.Unmarshal(out, &r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusNotFound, r.Extensions.Status)
	assert.Equal(t, "Not Found", r.Extensions.HTTPError)
	assert.Equal(t, "no peer ID available. P2P might be disabled", r.Message)
}

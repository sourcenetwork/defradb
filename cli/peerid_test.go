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

	"github.com/sourcenetwork/defradb/core/api"
	"github.com/stretchr/testify/assert"
)

func TestGetPeerIDCmd(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	cfg.Datastore.Store = "memory"
	cfg.Datastore.Badger.Path = dir
	di, err := start(ctx)
	if err != nil {
		t.Fatal(err)
	}

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

	r := api.DataResponse{}
	err = json.Unmarshal(out, &r)
	if err != nil {
		t.Fatal(err)
	}

	switch v := r.Data.(type) {
	case map[string]interface{}:
		assert.Equal(t, di.node.PeerID().String(), v["peerID"])
	default:
		t.Fatalf("data should be of type map[string]interface{} but got %T", r.Data)
	}

	di.close(ctx)
}

func TestGetPeerIDCmdWithNoP2P(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	cfg.Datastore.Store = "memory"
	cfg.Datastore.Badger.Path = dir
	cfg.Net.P2PDisabled = true
	di, err := start(ctx)
	if err != nil {
		t.Fatal(err)
	}

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

	r := api.ErrorResponse{}
	err = json.Unmarshal(out, &r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusNotFound, r.Errors[0].Extensions.Status)
	assert.Equal(t, "Not Found", r.Errors[0].Extensions.HTTPError)
	assert.Equal(t, "no peer ID available. P2P might be disabled", r.Errors[0].Message)

	di.close(ctx)
}

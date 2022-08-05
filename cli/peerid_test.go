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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPeerIDCmd(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	cfg.Datastore.Store = "memory"
	cfg.Datastore.Badger.Path = dir
	n, err := start(ctx)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)

	rootCmd.SetArgs([]string{"client", "peerid"})

	err = rootCmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	out, err := io.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}

	type peerIDResponse struct {
		Data struct {
			PeerID string `json:"peerID"`
		} `json:"data"`
	}
	r := peerIDResponse{}
	err = json.Unmarshal(out, &r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n.PeerID().String(), r.Data.PeerID)
}

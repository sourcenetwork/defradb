// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build !js

package tests

import (
	"fmt"

	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/tests/clients"
	cwrap "github.com/sourcenetwork/defradb/tests/clients/c"
	"github.com/sourcenetwork/defradb/tests/clients/cli"
	"github.com/sourcenetwork/defradb/tests/clients/http"
	"github.com/sourcenetwork/defradb/tests/state"
)

func init() {
	if !goClient && !httpClient && !cliClient && !cClient {
		// Default is to test go client type.
		goClient = true
	}
	if cClient {
		// TODO: Network test support for C client
		skipNetworkTests = true
		skipBackupTests = true
	}
}

// setupClient returns the client implementation for the current
// testing state. The client type on the test state is used to
// select the client implementation to use.
func setupClient(s *state.State, nodeObj *node.Node, enableNAC bool) (clients.Client, error) {
	switch s.ClientType {
	case HTTPClientType:
		return http.NewWrapper(nodeObj)

	case CLIClientType:
		return cli.NewWrapper(nodeObj, s.SourcehubAddress)

	case GoClientType:
		return newGoClientWrapper(nodeObj), nil

	case CClientType:
		return cwrap.NewCWrapper(s.Ctx, enableNAC), nil

	default:
		return nil, fmt.Errorf("invalid client type: %v", s.ClientType)
	}
}

type goClientWrapper struct {
	node.DB
	node.Peer
}

func newGoClientWrapper(n *node.Node) *goClientWrapper {
	return &goClientWrapper{
		DB:   n.DB,
		Peer: n.Peer,
	}
}

func (w *goClientWrapper) Close() {
	if w.Peer != nil {
		w.Peer.Close()
	}
	w.DB.Close()
}

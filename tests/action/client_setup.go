// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

//go:build !js

package action

import (
	"context"
	"fmt"

	cbindings "github.com/sourcenetwork/defradb/cbindings"
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/tests/clients"
	"github.com/sourcenetwork/defradb/tests/clients/cli"
	"github.com/sourcenetwork/defradb/tests/clients/http"
	"github.com/sourcenetwork/defradb/tests/state"
)

// setupClient returns the client implementation for the current
// testing state. The client type on the test state is used to
// select the client implementation to use.
func setupClient(s *state.State, nodeObj *node.Node) (clients.Client, error) {
	switch s.ClientType {
	case state.HTTPClientType:
		return http.NewWrapper(nodeObj)

	case state.CLIClientType:
		return cli.NewWrapper(nodeObj, s.RemoteDACAddress)

	case state.GoClientType:
		return newGoClientWrapper(nodeObj), nil

	case state.CClientType:
		return cbindings.NewCWrapper(nodeObj)

	default:
		return nil, fmt.Errorf("invalid client type: %v", s.ClientType)
	}
}

type goClientWrapper struct {
	node.DB
	node *node.Node
}

func newGoClientWrapper(n *node.Node) *goClientWrapper {
	return &goClientWrapper{
		DB:   n.DB,
		node: n,
	}
}

func (w *goClientWrapper) Close() {
	_ = w.node.Close(context.Background())
}

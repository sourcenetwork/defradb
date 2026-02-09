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
	"context"
	"fmt"

	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	cbindings "github.com/sourcenetwork/defradb/cbindings"
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/tests/clients"
	"github.com/sourcenetwork/defradb/tests/clients/cli"
	"github.com/sourcenetwork/defradb/tests/clients/http"
	"github.com/sourcenetwork/defradb/tests/state"
)

func init() {
	if !goClient && !httpClient && !cliClient && !cClient && !rustFFIClient {
		// Default is to test go client type.
		goClient = true
	}
	if cClient {
		skipNetworkTests = false
		skipBackupTests = true
	}
	if rustFFIClient {
		// Don't skip any tests - let them fail so we can track progress
		skipNetworkTests = false
		skipBackupTests = false
	}
}

// setupClient returns the client implementation for the current
// testing state. The client type on the test state is used to
// select the client implementation to use.
//
// The identity parameter is the identity used during Go node startup.
// For the Rust FFI client, this is used to mirror NAC initialization
// so the Rust node has the same access control state as the Go node.
func setupClient(
	s *state.State,
	nodeObj *node.Node,
	identity immutable.Option[acpIdentity.Identity],
	enableSigning bool,
	dbPath string,
) (clients.Client, error) {
	switch s.ClientType {
	case state.HTTPClientType:
		return http.NewWrapper(nodeObj)

	case state.CLIClientType:
		return cli.NewWrapper(nodeObj, s.SourcehubAddress)

	case state.GoClientType:
		return newGoClientWrapper(nodeObj), nil

	case state.CClientType:
		return cbindings.NewCWrapper(nodeObj)

	case state.RustFFIClientType:
		return setupRustFFIClient(s, nodeObj, identity, s.CurrentSetupNodeID, enableSigning, dbPath)

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

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

package tests

import (
	"context"
	"fmt"

	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	cbindings "github.com/sourcenetwork/defradb/cbindings"
	prodHttp "github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/tests/clients"
	"github.com/sourcenetwork/defradb/tests/clients/cli"
	"github.com/sourcenetwork/defradb/tests/clients/http"
	"github.com/sourcenetwork/defradb/tests/clients/rustffi"
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
) (clients.Client, error) {
	// The test suite completely bypasses the way production consumes the node options,
	// including the configuration of IsDevMode, so we have to hard code it here for now.
	prodHttp.IsDevMode = true


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
		return setupRustFFIClient(s, nodeObj, identity)

	default:
		return nil, fmt.Errorf("invalid client type: %v", s.ClientType)
	}
}

// setupRustFFIClient creates a Rust FFI wrapper and mirrors the Go node's NAC
// state onto it. This ensures the Rust FFI node has the same access control
// configuration as the Go node, matching Go's initializeNodeACP() flow.
func setupRustFFIClient(
	s *state.State,
	nodeObj *node.Node,
	identity immutable.Option[acpIdentity.Identity],
) (*rustffi.Wrapper, error) {
	var wrapper *rustffi.Wrapper
	var err error

	if s.IsNetworkEnabled {
		listenAddr := "/ip4/" + getIPString() + "/tcp/0"
		wrapper, err = rustffi.NewWrapperWithP2P(listenAddr)
	} else {
		wrapper, err = rustffi.NewWrapper()
	}
	if err != nil {
		return nil, err
	}

	// Mirror the Go node's NAC state onto the Rust FFI node.
	// The Go node may have NAC enabled (via WithEnableNodeACP + identity in context
	// during Start). The Rust FFI node starts fresh without NAC, so we check the
	// Go node's status and enable NAC on the Rust node with the same owner identity.
	nacStatus, nacErr := nodeObj.DB.GetNACStatus(s.Ctx)
	if nacErr == nil && nacStatus.Status == "enabled" && identity.HasValue() {
		ownerDID := identity.Value().DID()
		if err := wrapper.EnableNACForInit(ownerDID); err != nil {
			wrapper.Close()
			return nil, fmt.Errorf("failed to enable NAC on Rust FFI node: %w", err)
		}
	}

	return wrapper, nil
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

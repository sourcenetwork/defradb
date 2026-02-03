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
	nodeIndex int,
	enableSigning bool,
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
		return setupRustFFIClient(s, nodeObj, identity, nodeIndex, enableSigning)

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
	nodeIndex int,
	enableSigning bool,
) (*rustffi.Wrapper, error) {
	var wrapper *rustffi.Wrapper
	var err error

	// Get the node identity for this node index - this is the same identity
	// used for signing in the Go node (set via db.WithNodeIdentity)
	var nodeIdentity acpIdentity.Identity
	if enableSigning {
		nodeIdentity = state.GetIdentity(s, NodeIdentity(nodeIndex))
	}

	if s.IsNetworkEnabled {
		listenAddr := "/ip4/" + getIPString() + "/tcp/0"
		wrapper, err = rustffi.NewWrapperWithP2P(listenAddr, enableSigning, nodeIdentity)
	} else {
		wrapper, err = rustffi.NewWrapper(enableSigning, nodeIdentity)
	}
	if err != nil {
		return nil, err
	}

	// Forward SE artifact events from Go's event bus to the wrapper's event bus.
	// Go's SE coordinator publishes to Go's event bus, but the test framework
	// subscribes to the wrapper's event bus via c.Events().
	if err := wrapper.ForwardSEEvents(nodeObj.DB.Events()); err != nil {
		wrapper.Close()
		return nil, fmt.Errorf("failed to forward SE events: %w", err)
	}

	// Mirror the Go node's NAC state onto the Rust FFI node.
	//
	// The Go node has two NAC authorization paths (see checkNodeAccess in db_nac.go):
	// 1. nodeIdentity shortcut: if caller DID == db.nodeIdentity DID, access granted
	// 2. ACP owner check: the identity from context during Start() creates the NAC
	//    policy and becomes the owner in the Zanzibar store
	//
	// We mirror this by:
	// - Enabling NAC with the context identity (action.Identity) as owner
	// - Adding NodeIdentity as admin (mirrors the nodeIdentity shortcut)
	nacStatus, nacErr := nodeObj.DB.GetNACStatus(s.Ctx)
	if nacErr == nil && nacStatus.Status == "enabled" && identity.HasValue() {
		ownerDID := identity.Value().DID()
		if ownerDID != "" {
			if err := wrapper.EnableNACForInit(ownerDID); err != nil {
				wrapper.Close()
				return nil, fmt.Errorf("failed to enable NAC on Rust FFI node: %w", err)
			}
			// Mirror Go's nodeIdentity shortcut: the Go node grants automatic
			// access to db.nodeIdentity (set via WithNodeIdentity). Add that
			// identity as admin on the Rust FFI node so refreshCollections works.
			nodeIdentityDID := getIdentityDID(s, NodeIdentity(nodeIndex))
			if nodeIdentityDID != "" && nodeIdentityDID != ownerDID {
				if err := wrapper.AddNACAdminForInit(ownerDID, nodeIdentityDID); err != nil {
					wrapper.Close()
					return nil, fmt.Errorf("failed to add node identity as NAC admin: %w", err)
				}
			}
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

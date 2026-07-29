// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build !js && rust_ffi

package tests

import (
	"context"
	"fmt"

	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client/options"
	internalIdentity "github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/tests/clients"
	"github.com/sourcenetwork/defradb/tests/clients/rustffi"
	"github.com/sourcenetwork/defradb/tests/state"
)

// setupRustFFIClient creates a Rust FFI wrapper and mirrors the Go node's NAC
// state onto it. This ensures the Rust FFI node has the same access control
// configuration as the Go node, matching Go's initializeNodeACP() flow.
func setupRustFFIClient(
	s *state.State,
	nodeObj *node.Node,
	identity immutable.Option[acpIdentity.Identity],
	nodeIndex int,
	enableSigning bool,
	dbPath string,
) (clients.Client, error) {
	var wrapper *rustffi.Wrapper
	var err error

	// Get the node identity for this node index. Go always configures a node
	// identity (used for GetNodeIdentity, P2P peer ID, and optionally signing).
	// The Rust FFI node needs this regardless of enableSigning so that
	// GetNodeIdentity returns the correct DID.
	nodeIdentity := state.GetIdentity(s, NodeIdentity(nodeIndex))

	// For file-based storage, use a subdirectory so Rust's redb doesn't collide
	// with Go's badger store at the same path.
	rustDBPath := ""
	if dbPath != "" {
		rustDBPath = dbPath + "/rustffi"
	}

	// Build SourceHub config if SourceHub ACP is active.
	// GRPCAddress here is actually the REST/LCD API address (the Rust client uses REST, not gRPC).
	var shConfig *rustffi.SourceHubConfig
	if s.DocumentACPType == state.SourceHubDocumentACPType && s.SourcehubRestAddress != "" {
		shConfig = &rustffi.SourceHubConfig{
			GRPCAddress:     s.SourcehubRestAddress,
			CometRPCAddress: s.SourcehubCometRPCAddress,
			ChainID:         s.SourcehubChainID,
			SignerKey:       s.SourcehubSignerKey,
		}
	}

	if s.IsNetworkEnabled {
		listenAddr := "/ip4/" + getIPString() + "/tcp/0"
		wrapper, err = rustffi.NewWrapperWithP2P(listenAddr, enableSigning, nodeIdentity, rustDBPath, shConfig)
	} else {
		wrapper, err = rustffi.NewWrapper(enableSigning, nodeIdentity, rustDBPath, shConfig)
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

	// Pass the SE encryption key to the Rust FFI node so it can generate
	// SE artifacts during replication (push_existing_docs).
	type seKeyProvider interface {
		SearchableEncryptionKey() []byte
	}
	if skp, ok := nodeObj.DB.(seKeyProvider); ok {
		if seKey := skp.SearchableEncryptionKey(); len(seKey) > 0 {
			if err := wrapper.SetSEEncryptionKey(seKey); err != nil {
				wrapper.Close()
				return nil, fmt.Errorf("failed to set SE encryption key on Rust FFI node: %w", err)
			}
		}
	}

	// Mirror the Go node's NAC state onto the Rust FFI node.
	// Pass identity via options (not just context) because Go's GetNACStatus
	// uses opt.Identity for the NAC permission check, not the context identity.
	var identityCtx context.Context
	if identity.HasValue() {
		identityCtx = internalIdentity.WithContext(s.Ctx, identity)
	} else {
		identityCtx = s.Ctx
	}
	getNACOpt := options.WithIdentity(options.GetNACStatus(), identity)
	nacStatus, nacErr := nodeObj.DB.GetNACStatus(identityCtx, getNACOpt)
	if nacErr == nil && identity.HasValue() {
		ownerDID := identity.Value().DID()
		if ownerDID != "" && (nacStatus.Status == "enabled" || nacStatus.Status == "disabled temporarily") {
			// Check if Rust node already has NAC enabled (from persistent store).
			// Pass identity as an option so the FFI NAC permission check succeeds.
			rustNACOpt := options.WithIdentity(options.GetNACStatus(), identity)
			rustStatus, rustErr := wrapper.GetNACStatus(identityCtx, rustNACOpt)
			needsEnable := rustErr != nil || rustStatus.Status == "not configured"

			if needsEnable {
				if err := wrapper.EnableNACForInit(ownerDID); err != nil {
					wrapper.Close()
					return nil, fmt.Errorf("failed to enable NAC on Rust FFI node: %w", err)
				}
			}
			// Mirror Go's nodeIdentity shortcut
			nodeIdentityDID := state.GetIdentityDID(s, NodeIdentity(nodeIndex))
			rustStatusForAdmin, _ := wrapper.GetNACStatus(identityCtx, rustNACOpt)
			if nodeIdentityDID != "" && nodeIdentityDID != ownerDID && rustStatusForAdmin.Status != "disabled temporarily" {
				if err := wrapper.AddNACAdminForInit(ownerDID, nodeIdentityDID); err != nil {
					wrapper.Close()
					return nil, fmt.Errorf("failed to add node identity as NAC admin: %w", err)
				}
			}

			// If Go says "disabled temporarily", mirror that state on Rust
			if nacStatus.Status == "disabled temporarily" {
				rustStatus2, rustErr2 := wrapper.GetNACStatus(identityCtx, rustNACOpt)
				if rustErr2 != nil || rustStatus2.Status != "disabled temporarily" {
					ctx := internalIdentity.WithContext(s.Ctx, identity)
					if err := wrapper.DisableNAC(ctx); err != nil {
						wrapper.Close()
						return nil, fmt.Errorf("failed to mirror NAC disabled state: %w", err)
					}
				}
			}
		}
	}

	// Store the Go node closer so the Go node's badger lock is released
	// when the wrapper is closed during restart tests.
	wrapper.SetGoNodeCloser(func() {
		_ = nodeObj.Close(context.Background())
	})

	return wrapper, nil
}

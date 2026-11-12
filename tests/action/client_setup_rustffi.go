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

//go:build !js && rust_ffi

package action

import (
	"context"
	"fmt"

	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/tests/clients"
	"github.com/sourcenetwork/defradb/tests/clients/rustffi"
	"github.com/sourcenetwork/defradb/tests/state"
)

// setupRustFFIClient starts a Rust FFI node configured as the Go node next to
// it was, and returns it as the test's client.
//
// The Go node stays open, without P2P, so the harness can still read its
// configuration; the wrapper closes it when it closes.
func setupRustFFIClient(
	s *state.State,
	nodeObj *node.Node,
	startIdentity immutable.Option[state.Identity],
	cfg NodeSetupConfig,
	path string,
) (clients.Client, error) {
	nodeIdentity := state.GetIdentity(s, NodeIdentity(s.CurrentSetupNodeID))
	starter := resolveIdentity(s, startIdentity)

	rustCfg := rustffi.Config{
		EnableSigning:            cfg.EnableSigning,
		NodeIdentity:             nodeIdentity,
		ReplicatorRetryIntervals: "1",
	}

	// Go's store sits at path, so the Rust store takes a directory of its own.
	if path != "" {
		rustCfg.DBPath = path + "/rustffi"
	}

	if s.IsNetworkEnabled && !cfg.DisableP2P {
		rustCfg.ListenAddresses = rustListenAddresses(s)
	}

	type seKeyProvider interface {
		SearchableEncryptionKey() []byte
	}
	if provider, ok := nodeObj.DB.(seKeyProvider); ok {
		rustCfg.SearchableEncryptionKey = provider.SearchableEncryptionKey()
	}
	if cfg.BadgerEncryption {
		rustCfg.StoreEncryptionKey = encryptionKey
	}

	if s.DocumentACPType == state.RemoteDocumentACPType && s.DocumentACPOptions != nil {
		dac := s.DocumentACPOptions
		rustCfg.RemoteDAC = &rustffi.RemoteDACConfig{
			GRPCAddress:     dac.RemoteDACGRPCAddress,
			CometRPCAddress: dac.RemoteDACCometRPCAddress,
			ChainID:         dac.RemoteDACLogID,
			SignerKey:       getAccountDataFromMnemonic(s.T, faucetMnemonic),
		}
	}

	nacStatus := goNACStatus(s, nodeObj, starter)
	rustCfg.EnableNAC = nacStatus == client.NACEnabled.String() ||
		nacStatus == client.NACDisabledTemporarily.String()

	wrapper, err := rustffi.NewWrapper(rustCfg)
	if err != nil {
		return nil, err
	}
	wrapper.SetGoNodeCloser(func() {
		_ = nodeObj.Close(context.Background())
	})

	if err := mirrorNAC(wrapper, nacStatus, nodeIdentity, starter); err != nil {
		wrapper.Close()
		return nil, err
	}
	return wrapper, nil
}

// rustListenAddresses keeps a restarted node on its old addresses, because its
// peers keep dialling those, and gives a new node any free port.
func rustListenAddresses(s *state.State) []string {
	id := s.CurrentSetupNodeID
	if id < len(s.Nodes) && s.Nodes[id] != nil && len(s.Nodes[id].CachedAddresses) > 0 {
		return s.Nodes[id].CachedAddresses
	}
	return []string{"/ip4/" + getIPString() + "/tcp/0"}
}

// goNACStatus reads the Go node's access control status as the identity that
// started it, the only one NAC lets read it. It is empty when NAC is off.
func goNACStatus(
	s *state.State,
	nodeObj *node.Node,
	starter immutable.Option[acpIdentity.Identity],
) string {
	if !starter.HasValue() {
		return ""
	}
	opt := options.WithIdentity(options.GetNACStatus(), starter)
	status, err := nodeObj.DB.GetNACStatus(s.Ctx, opt)
	if err != nil {
		return ""
	}
	return status.Status
}

// mirrorNAC gives the Rust node the access control the Go node has.
//
// Go makes the identity that starts the node its NAC owner and lets the node
// identity through as well. The Rust node is owned by its node identity, so
// the starting identity is made an admin of it instead.
func mirrorNAC(
	wrapper *rustffi.Wrapper,
	goStatus string,
	nodeIdentity acpIdentity.Identity,
	starter immutable.Option[acpIdentity.Identity],
) error {
	nacOn := goStatus == client.NACEnabled.String() ||
		goStatus == client.NACDisabledTemporarily.String()
	if !nacOn || !starter.HasValue() {
		return nil
	}
	if nodeIdentity == nil || starter.Value().DID() != nodeIdentity.DID() {
		if err := wrapper.GrantNACAdmin(starter.Value()); err != nil {
			return fmt.Errorf("making the starting identity a NAC admin: %w", err)
		}
	}
	if goStatus == client.NACDisabledTemporarily.String() {
		if err := wrapper.DisableNACAsNode(); err != nil {
			return fmt.Errorf("mirroring NAC disabled temporarily: %w", err)
		}
	}
	return nil
}

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

package action

import (
	"time"

	"github.com/stretchr/testify/require"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
	changeDetector "github.com/sourcenetwork/defradb/tests/change_detector"
	"github.com/sourcenetwork/defradb/tests/state"
)

// ConfigureNode returns the P2P options for a new Defra node.
type ConfigureNode func() options.NodeP2POptions

// NewNode creates a new Defra node. The zero value creates a native,
// current-build node with default networking.
//
// If no nodes are explicitly configured, a default one will be setup.  There is no
// upper limit to the number that can be configured.
//
// Nodes may be explicitly referenced by index by other actions using `NodeID` properties.
// If the action has a `NodeID` property and it is not specified, the action will be
// effected on all nodes.
//
// Configuration is held as plain data so a multiplier can rewrite it to run existing
// tests under other node configurations.
type NewNode struct {
	stateful

	// Version, when set (e.g. "v1.0.0"), runs the node as an external process from
	// that published release binary instead of natively in-process.
	Version string
	// Network returns the node's P2P options. Nil means default networking.
	// Ignored if DisableP2P is true.
	Network ConfigureNode
	// DisableP2P, when true, starts the node with P2P disabled entirely (the node's
	// internal db.p2p stays nil), instead of the default networking Network configures.
	DisableP2P bool

	// SetupConfig carries the test-level settings node setup needs. The harness
	// sets it before execution.
	SetupConfig NodeSetupConfig
}

var _ Action = (*NewNode)(nil)
var _ Stateful = (*NewNode)(nil)

// P2POptions returns the configured P2P options, or the defaults if no networking
// config was supplied.
func (a *NewNode) P2POptions() options.NodeP2POptions {
	if a.Network == nil {
		return options.NodeP2POptions{}
	}
	return a.Network()
}

// WithVersion returns a copy of the action that runs the node as an external
// process from the given published release, e.g. "v1.0.0".
//
// The copy is shallow: the Network function and the fields of SetupConfig are
// shared with the original, not cloned.
func (a *NewNode) WithVersion(version string) *NewNode {
	clone := *a
	clone.Version = version
	return &clone
}

// Execute configures and starts a new Defra node.
//
// Any errors generated during configuration will result in a test failure.
func (a *NewNode) Execute() {
	s := a.s

	if changeDetector.Enabled {
		// We do not yet support the change detector for tests running across multiple nodes.
		s.T.SkipNow()
		return
	}

	p2pOpts := a.P2POptions()
	s.CurrentSetupNodeID = len(s.Nodes)

	// Versioned nodes run in a separate process from a release binary that configures
	// itself, so in-process options do not apply to them.
	var opts *options.NodeOptionsBuilder
	if a.Version == "" {
		privateKey, err := crypto.GenerateEd25519()
		require.NoError(s.T, err)
		WithPrivateKey(&p2pOpts, privateKey)

		opts = DefaultNodeOpts(a.SetupConfig)
		opts.DB().
			SetRetryIntervals([]time.Duration{time.Millisecond * 1}).
			SetNodeIdentity(state.GetIdentity(s, NodeIdentity(s.CurrentSetupNodeID)))
		opts.P2P().SetAll(p2pOpts)
		opts.SetDisableP2P(a.DisableP2P)
	}

	node, err := SetupNode(s, acpIdentity.None, a.SetupConfig, opts, a.Version)
	require.NoError(s.T, err)
	if node == nil {
		// SetupNode already skipped the test (no release asset for this platform).
		return
	}

	node.P2POpts = p2pOpts
	node.DisableP2P = a.DisableP2P
	node.Version = a.Version
	s.Nodes = append(s.Nodes, node)
}

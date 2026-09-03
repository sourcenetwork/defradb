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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/state"
)

const (
	BadgerIMType   state.DatabaseType = "badger-in-memory"
	DefraIMType    state.DatabaseType = "defra-memory-datastore"
	BadgerFileType state.DatabaseType = "badger-file-system"
	LevelStoreType state.DatabaseType = "level"
)

const (
	// NoneKMSType is the none KMS type. It is used to indicate that no KMS should be used.
	NoneKMSType state.KMSType = "none"
	// PubSubKMSType is the PubSub KMS type.
	PubSubKMSType state.KMSType = "pubsub"
)

// NodeSetupConfig carries the test-level settings node setup needs.
//
// Node setup reads only these few values, so taking them directly keeps it
// independent of the test case type.
type NodeSetupConfig struct {
	// EnableSigning enables document signing on the node.
	EnableSigning bool
	// HTTP overrides the node's HTTP server settings when set.
	HTTP immutable.Option[options.NodeHTTPOptions]
	// IsDocumentACPTest reports whether the test uses document ACP, which
	// decides whether a Vera instance is needed at all.
	IsDocumentACPTest bool
	// VeraImage is the container image used to run Vera.
	VeraImage string
	// DatabaseDir, when set, is the path a restarting node reopens its store
	// from. Empty means a fresh directory.
	DatabaseDir string
	// BadgerEncryption enables encryption on a badger store.
	BadgerEncryption bool
	// LensRuntime selects the lens runtime, and LensPoolSize how many are
	// instantiated. A zero pool size leaves the node default in place.
	LensRuntime  options.NodeLensRuntimeType
	LensPoolSize int
}

func applyHTTPOptions(opts *options.NodeOptionsBuilder, httpOpts options.NodeHTTPOptions) {
	httpBuilder := opts.HTTP()
	if httpOpts.Address != "" {
		httpBuilder.SetAddress(httpOpts.Address)
	}
	if len(httpOpts.AllowedOrigins) > 0 {
		httpBuilder.SetAllowedOrigins(httpOpts.AllowedOrigins...)
	}
	if httpOpts.TLSCertPath != "" {
		httpBuilder.SetCertPath(httpOpts.TLSCertPath)
	}
	if httpOpts.TLSKeyPath != "" {
		httpBuilder.SetKeyPath(httpOpts.TLSKeyPath)
	}
	if httpOpts.ReadTimeout != 0 {
		httpBuilder.SetReadTimeout(httpOpts.ReadTimeout)
	}
	if httpOpts.WriteTimeout != 0 {
		httpBuilder.SetWriteTimeout(httpOpts.WriteTimeout)
	}
	if httpOpts.IdleTimeout != 0 {
		httpBuilder.SetIdleTimeout(httpOpts.IdleTimeout)
	}
	if httpOpts.TxnTTL != 0 {
		httpBuilder.SetTxnTTL(httpOpts.TxnTTL)
	}
	if httpOpts.TxnTTLTick != 0 {
		httpBuilder.SetTxnTTLTick(httpOpts.TxnTTLTick)
	}
	if httpOpts.TxnTTLBuckets != 0 {
		httpBuilder.SetTxnTTLBuckets(httpOpts.TxnTTLBuckets)
	}
}

// DefaultNodeOpts returns the node options shared by every test node.
func DefaultNodeOpts(cfg NodeSetupConfig) *options.NodeOptionsBuilder {
	opt := options.Node().
		// The test framework sets this up elsewhere when required so that it may be wrapped
		// into a [client.TxnStore].
		SetDisableAPI(true).
		// The p2p is configured in the tests by [NewNode] actions, we disable it here
		// to keep the tests as lightweight as possible.
		SetDisableP2P(true)

	if cfg.LensPoolSize != 0 {
		opt.DB().SetLensPoolSize(cfg.LensPoolSize)
	}
	opt.DB().
		SetLensRuntime(cfg.LensRuntime).
		// The default is 5 and that is never going to be needed in a testing scenario where all the
		// nodes are on the same machine with no network latency.
		SetP2PBlockSyncTimeout(1 * time.Second)

	return opt
}

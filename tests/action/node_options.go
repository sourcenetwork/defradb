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
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/options"
	changeDetector "github.com/sourcenetwork/defradb/tests/change_detector"
	"github.com/sourcenetwork/defradb/tests/state"
)

const (
	memoryBadgerEnvName     = "DEFRA_BADGER_MEMORY"
	fileBadgerEnvName       = "DEFRA_BADGER_FILE"
	badgerEncryptionEnvName = "DEFRA_BADGER_ENCRYPTION"
	levelEnvName            = "DEFRA_LEVEL"
	inMemoryEnvName         = "DEFRA_IN_MEMORY"
	lensTypeEnvName         = "DEFRA_LENS_TYPE"

	lensPoolSize = 2
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
	// decides whether a SourceHub instance is needed at all.
	IsDocumentACPTest bool
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

var (
	// BadgerInMemory, BadgerFile, InMemoryStore and LevelStore select the store
	// types under test. Node setup and the test harness both read them, so they
	// are resolved once here rather than copied into each package.
	BadgerInMemory bool
	BadgerFile     bool
	InMemoryStore  bool
	LevelStore     bool

	// DatabaseDir is the path a restarting node reopens its store from. It is
	// set by the harness around restart actions.
	DatabaseDir string

	// lensType is the lens runtime under test.
	lensType options.NodeLensRuntimeType

	// BadgerEncryption reports whether the badger store is encrypted.
	BadgerEncryption bool

	encryptionKey []byte
	// encryptionKeyOnce guards the lazy, process-wide initialization of
	// encryptionKey so concurrent node setups don't race on it.
	encryptionKeyOnce sync.Once
	encryptionKeyErr  error
)

func init() {
	// We use environment variables instead of flags `go test ./...` throws for all packages
	// that don't have the flag defined
	BadgerFile, _ = strconv.ParseBool(os.Getenv(fileBadgerEnvName))
	BadgerInMemory, _ = strconv.ParseBool(os.Getenv(memoryBadgerEnvName))
	InMemoryStore, _ = strconv.ParseBool(os.Getenv(inMemoryEnvName))
	LevelStore, _ = strconv.ParseBool(os.Getenv((levelEnvName)))
	BadgerEncryption, _ = strconv.ParseBool(os.Getenv(badgerEncryptionEnvName))
	lensType = options.NodeLensRuntimeType(os.Getenv(lensTypeEnvName))

	if changeDetector.Enabled {
		// Change detector only uses badger file db type.
		BadgerFile = true
		BadgerInMemory = false
		InMemoryStore = false
		LevelStore = false
	} else if !BadgerInMemory && !BadgerFile && !InMemoryStore && !LevelStore {
		// Default is to test all but filesystem db types.
		BadgerFile = false
		BadgerInMemory = true
		InMemoryStore = false
		LevelStore = false
	}
}

// DefaultNodeOpts returns the node options shared by every test node.
func DefaultNodeOpts() *options.NodeOptionsBuilder {
	opt := options.Node().
		// The test framework sets this up elsewhere when required so that it may be wrapped
		// into a [client.TxnStore].
		SetDisableAPI(true).
		// The p2p is configured in the tests by [NodeConfig] actions, we disable it here
		// to keep the tests as lightweight as possible.
		SetDisableP2P(true)

	opt.DB().
		SetLensPoolSize(lensPoolSize).
		SetLensRuntime(lensType).
		// The default is 5 and that is never going to be needed in a testing scenario where all the
		// nodes are on the same machine with no network latency.
		SetP2PBlockSyncTimeout(1 * time.Second)

	return opt
}

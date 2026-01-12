// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tests

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/node"
	changeDetector "github.com/sourcenetwork/defradb/tests/change_detector"
	"github.com/sourcenetwork/defradb/tests/state"
)

// NodeSetupOptions contains options for setting up a node in tests.
type NodeSetupOptions struct {
	// P2POpts contains P2P configuration options.
	P2POpts options.NodeP2POptions
	// NodeIdentity is the identity to use for the node.
	NodeIdentity acpIdentity.Identity
	// EnableNAC enables Node Access Control.
	EnableNAC bool
	// RetryIntervals specifies replicator retry intervals.
	RetryIntervals []time.Duration
}

const (
	memoryBadgerEnvName     = "DEFRA_BADGER_MEMORY"
	fileBadgerEnvName       = "DEFRA_BADGER_FILE"
	fileBadgerPathEnvName   = "DEFRA_BADGER_FILE_PATH"
	badgerEncryptionEnvName = "DEFRA_BADGER_ENCRYPTION"
	inMemoryEnvName         = "DEFRA_IN_MEMORY"
)

const (
	BadgerIMType   state.DatabaseType = "badger-in-memory"
	DefraIMType    state.DatabaseType = "defra-memory-datastore"
	BadgerFileType state.DatabaseType = "badger-file-system"
)

var (
	badgerInMemory   bool
	badgerFile       bool
	inMemoryStore    bool
	databaseDir      string
	badgerEncryption bool
	encryptionKey    []byte
)

func init() {
	// We use environment variables instead of flags `go test ./...` throws for all packages
	// that don't have the flag defined
	badgerFile, _ = strconv.ParseBool(os.Getenv(fileBadgerEnvName))
	badgerInMemory, _ = strconv.ParseBool(os.Getenv(memoryBadgerEnvName))
	inMemoryStore, _ = strconv.ParseBool(os.Getenv(inMemoryEnvName))
	badgerEncryption, _ = strconv.ParseBool(os.Getenv(badgerEncryptionEnvName))

	if changeDetector.Enabled {
		// Change detector only uses badger file db type.
		badgerFile = true
		badgerInMemory = false
		inMemoryStore = false
	} else if !badgerInMemory && !badgerFile && !inMemoryStore {
		// Default is to test all but filesystem db types.
		badgerFile = false
		badgerInMemory = true
		inMemoryStore = false
	}
}

func defaultNodeOpts() *options.NodeOptions {
	nodeOpts := options.Node()
	// The test framework sets this up elsewhere when required so that it may be wrapped
	// into a [client.TxnStore].
	nodeOpts.DisableAPI = true
	// The p2p is configured in the tests by [ConfigureNode] actions, we disable it here
	// to keep the tests as lightweight as possible.
	nodeOpts.DisableP2P = true
	nodeOpts.DB.LensPoolSize = lensPoolSize
	nodeOpts.DB.LensRuntime = options.NodeLensRuntimeType(lensType)
	// The default is 5 and that is never going to be needed in a testing scenario where all the
	// nodes are on the same machine with no network latency.
	nodeOpts.DB.P2PBlockSyncTimeout = 1 * time.Second
	return nodeOpts
}

func NewBadgerMemoryDB(ctx context.Context) (node.DB, error) {
	nodeOpts := options.Node()
	nodeOpts.DisableP2P = true
	nodeOpts.DisableAPI = true
	nodeOpts.Store.BadgerInMemory = true

	n, err := node.New(ctx, nodeOpts)
	if err != nil {
		return nil, err
	}
	err = n.Start(ctx)
	if err != nil {
		return nil, err
	}
	return n.DB, err
}

func NewBadgerFileDB(ctx context.Context, t testing.TB) (node.DB, error) {
	path := t.TempDir()

	nodeOpts := options.Node()
	nodeOpts.DisableP2P = true
	nodeOpts.DisableAPI = true
	nodeOpts.Store.Path = path

	n, err := node.New(ctx, nodeOpts)
	if err != nil {
		return nil, err
	}
	err = n.Start(ctx)
	if err != nil {
		return nil, err
	}
	return n.DB, err
}

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

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/node"
	changeDetector "github.com/sourcenetwork/defradb/tests/change_detector"
	"github.com/sourcenetwork/defradb/tests/state"
)

const (
	memoryBadgerEnvName     = "DEFRA_BADGER_MEMORY"
	fileBadgerEnvName       = "DEFRA_BADGER_FILE"
	fileBadgerPathEnvName   = "DEFRA_BADGER_FILE_PATH"
	badgerEncryptionEnvName = "DEFRA_BADGER_ENCRYPTION"
	levelEnvName            = "DEFRA_LEVEL"
	inMemoryEnvName         = "DEFRA_IN_MEMORY"
)

const (
	BadgerIMType   state.DatabaseType = "badger-in-memory"
	DefraIMType    state.DatabaseType = "defra-memory-datastore"
	BadgerFileType state.DatabaseType = "badger-file-system"
	LevelStoreType state.DatabaseType = "level"
)

var (
	badgerInMemory   bool
	badgerFile       bool
	inMemoryStore    bool
	levelStore       bool
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
	levelStore, _ = strconv.ParseBool(os.Getenv((levelEnvName)))
	badgerEncryption, _ = strconv.ParseBool(os.Getenv(badgerEncryptionEnvName))

	if changeDetector.Enabled {
		// Change detector only uses badger file db type.
		badgerFile = true
		badgerInMemory = false
		inMemoryStore = false
		levelStore = false
	} else if !badgerInMemory && !badgerFile && !inMemoryStore && !levelStore {
		// Default is to test all but filesystem db types.
		badgerFile = false
		badgerInMemory = true
		inMemoryStore = false
		levelStore = false
	}
}

func defaultNodeOpts() *options.NodeOptionsBuilder {
	opt := options.Node().
		// The test framework sets this up elsewhere when required so that it may be wrapped
		// into a [client.TxnStore].
		SetDisableAPI(true).
		// The p2p is configured in the tests by [ConfigureNode] actions, we disable it here
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

func NewBadgerMemoryDB(ctx context.Context) (node.DB, error) {
	opts := options.Node().
		SetDisableP2P(true).
		SetDisableAPI(true).
		SetEnableDevelopment(true).
		Store().SetBadgerInMemory(true).
		Node()

	n, err := node.New(ctx, opts)
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

	opts := options.Node().
		SetDisableP2P(true).
		SetDisableAPI(true).
		SetEnableDevelopment(true).
		Store().SetPath(path).
		Node()

	n, err := node.New(ctx, opts)
	if err != nil {
		return nil, err
	}
	err = n.Start(ctx)
	if err != nil {
		return nil, err
	}
	return n.DB, err
}

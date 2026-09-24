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

package tests

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/tests/action"
	changeDetector "github.com/sourcenetwork/defradb/tests/change_detector"
)

const (
	memoryBadgerEnvName      = "DEFRA_BADGER_MEMORY"
	fileBadgerEnvName        = "DEFRA_BADGER_FILE"
	badgerEncryptionEnvName  = "DEFRA_BADGER_ENCRYPTION"
	levelEnvName             = "DEFRA_LEVEL"
	inMemoryEnvName          = "DEFRA_IN_MEMORY"
	lensTypeEnvName          = "DEFRA_LENS_TYPE"
	crossVersionExactEnvName = "DEFRA_CROSS_VERSION_EXACT"

	// Instantiating lenses is expensive, and our tests do not benefit from a large
	// number of them, so we explicitly set it to a low value.
	lensPoolSize = 2
)

const (
	BadgerIMType   = action.BadgerIMType
	DefraIMType    = action.DefraIMType
	BadgerFileType = action.BadgerFileType
	LevelStoreType = action.LevelStoreType
)

var (
	badgerInMemory bool
	badgerFile     bool
	inMemoryStore  bool
	levelStore     bool

	// databaseDir is the path a restarting node reopens its store from. It is
	// set by the harness around restart actions.
	databaseDir string

	// lensType is the lens runtime under test.
	lensType options.NodeLensRuntimeType

	// badgerEncryption reports whether the badger store is encrypted.
	badgerEncryption bool

	// crossVersionExact makes a version-targeting multiplier run only the release
	// it targets, skipping tests that need a newer one rather than moving them to
	// it. Runs covering several releases set this so a test is not run twice, once
	// by the release it needs and again by an older one promoting it.
	crossVersionExact bool
)

func init() {
	// We use environment variables instead of flags `go test ./...` throws for all packages
	// that don't have the flag defined
	badgerFile, _ = strconv.ParseBool(os.Getenv(fileBadgerEnvName))
	badgerInMemory, _ = strconv.ParseBool(os.Getenv(memoryBadgerEnvName))
	inMemoryStore, _ = strconv.ParseBool(os.Getenv(inMemoryEnvName))
	levelStore, _ = strconv.ParseBool(os.Getenv((levelEnvName)))
	badgerEncryption, _ = strconv.ParseBool(os.Getenv(badgerEncryptionEnvName))
	lensType = options.NodeLensRuntimeType(os.Getenv(lensTypeEnvName))
	crossVersionExact, _ = strconv.ParseBool(os.Getenv(crossVersionExactEnvName))

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

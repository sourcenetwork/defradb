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

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/defradb"
	"github.com/sourcenetwork/defradb/client"
	changeDetector "github.com/sourcenetwork/defradb/tests/change_detector"
)

type DatabaseType string

const (
	memoryBadgerEnvName     = "DEFRA_BADGER_MEMORY"
	fileBadgerEnvName       = "DEFRA_BADGER_FILE"
	fileBadgerPathEnvName   = "DEFRA_BADGER_FILE_PATH"
	badgerEncryptionEnvName = "DEFRA_BADGER_ENCRYPTION"
	inMemoryEnvName         = "DEFRA_IN_MEMORY"
)

const (
	BadgerIMType   DatabaseType = "badger-in-memory"
	DefraIMType    DatabaseType = "defra-memory-datastore"
	BadgerFileType DatabaseType = "badger-file-system"
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

func NewBadgerMemoryDB(ctx context.Context) (client.DB, error) {
	store := memory.NewDatastore(ctx)
	db, err := defradb.Open(ctx, store)
	if err != nil {
		return nil, err
	}
	return db, err
}

func NewBadgerFileDB(ctx context.Context, t testing.TB) (client.DB, error) {
	return NewBadgerMemoryDB(ctx)
}

// setupNode returns the database implementation for the current
// testing state. The database type on the test state is used to
// select the datastore implementation to use.
func setupNode(s *state, opts ...any) (*nodeState, error) {
	node, err := NewBadgerMemoryDB(s.ctx)
	if err != nil {
		return nil, err
	}

	c, err := setupClient(s, node)
	require.Nil(s.t, err)

	eventState, err := newEventState(c.Events())
	require.NoError(s.t, err)

	st := &nodeState{
		Client: c,
		event:  eventState,
		p2p:    newP2PState(),
	}

	return st, nil
}

// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"
	"testing"

	badgerds "github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/badger"
	"github.com/sourcenetwork/corekv/leveldb"

	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
)

// setForTest sets *p to v for the duration of the test, restoring the previous value on cleanup.
// It collapses the orig := X; X = v; defer func() { X = orig }() dance used to tune package vars
// (indexBackfillBatchSize, indexBuildConcurrency, indexBuildRetryDelay) into a single line.
func setForTest[T any](t *testing.T, p *T, v T) {
	t.Helper()
	orig := *p
	*p = v
	t.Cleanup(func() { *p = orig })
}

func newBadgerDB(ctx context.Context) (*DB, error) {
	rootstore, err := badger.NewDatastore("", badgerds.DefaultOptions("").WithInMemory(true))
	if err != nil {
		return nil, err
	}

	adminInfo, err := acpDB.NewNACInfo(ctx, "", false)
	if err != nil {
		return nil, err
	}
	return newDB(ctx, rootstore, adminInfo)
}

// newBadgerDBNoIndexWorker opens an in-memory DB whose index build worker is constructed but not
// started, so a test can drive draining via db.indexBuildWorker.drainSync without the background
// loop racing its hand-seeded records.
func newBadgerDBNoIndexWorker(t *testing.T, ctx context.Context) *DB {
	t.Helper()
	suppressIndexWorkerRun = true
	defer func() { suppressIndexWorkerRun = false }()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestNewDB(t *testing.T) {
	ctx := context.Background()
	rootstore, err := badger.NewDatastore("", badgerds.DefaultOptions("").WithInMemory(true))
	require.NoError(t, err)

	adminInfo, err := acpDB.NewNACInfo(ctx, "", false)
	require.NoError(t, err)

	_, err = NewDB(ctx, rootstore, adminInfo)
	require.NoError(t, err)
}

func TestReserveDocShortID_LevelDBUsesCurrentTxn(t *testing.T) {
	ctx := context.Background()
	rootstore, err := leveldb.NewDatastore(t.TempDir(), nil)
	require.NoError(t, err)

	adminInfo, err := acpDB.NewNACInfo(ctx, "", false)
	require.NoError(t, err)
	db, err := newDB(ctx, rootstore, adminInfo)
	require.NoError(t, err)
	t.Cleanup(db.Close)

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	defer txn.Discard()

	docShortID, err := db.reserveDocShortID(InitContext(ctx, txn))
	require.NoError(t, err)
	require.NotZero(t, docShortID)
	require.NoError(t, txn.Commit())
}

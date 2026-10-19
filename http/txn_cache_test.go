// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package http

import (
	"context"
	stdHTTP "net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client/options"
)

func testTxnCache(t *testing.T, txnTTL time.Duration) *txnCache {
	t.Helper()

	cache, err := newTxnCache(context.Background(), &options.NodeHTTPOptions{
		TxnTTL:        txnTTL,
		TxnTTLTick:    10 * time.Millisecond,
		TxnTTLBuckets: 20,
	})
	require.NoError(t, err)
	t.Cleanup(cache.Close)
	return cache
}

func TestTxnCache_HeldLeaseExpiresAfterRelease(t *testing.T) {
	cdb := setupDatabase(t)
	tx, err := cdb.NewTxn(false)
	require.NoError(t, err)

	cache := testTxnCache(t, 30*time.Millisecond)
	require.NoError(t, cache.Store(tx))

	lease, ok := cache.Acquire(tx.ID())
	require.True(t, ok)

	time.Sleep(90 * time.Millisecond)
	require.True(t, isTxnCached(cache, tx.ID()))

	lease.Release()

	require.Eventually(t, func() bool {
		return !isTxnCached(cache, tx.ID())
	}, time.Second, 10*time.Millisecond)
}

func TestTxnCache_ConcurrentLeasesExpireAfterLastRelease(t *testing.T) {
	cdb := setupDatabase(t)
	tx, err := cdb.NewTxn(false)
	require.NoError(t, err)

	cache := testTxnCache(t, 30*time.Millisecond)
	require.NoError(t, cache.Store(tx))

	firstLease, ok := cache.Acquire(tx.ID())
	require.True(t, ok)
	secondLease, ok := cache.Acquire(tx.ID())
	require.True(t, ok)

	firstLease.Release()
	time.Sleep(90 * time.Millisecond)
	require.True(t, isTxnCached(cache, tx.ID()))

	secondLease.Release()

	require.Eventually(t, func() bool {
		return !isTxnCached(cache, tx.ID())
	}, time.Second, 10*time.Millisecond)
}

func TestTxnCache_LoadAndDeletePreventsReleaseReschedule(t *testing.T) {
	cdb := setupDatabase(t)
	tx, err := cdb.NewTxn(false)
	require.NoError(t, err)

	cache := testTxnCache(t, 30*time.Millisecond)
	require.NoError(t, cache.Store(tx))

	lease, ok := cache.Acquire(tx.ID())
	require.True(t, ok)

	removedTxn, _, ok := cache.LoadAndDelete(tx.ID())
	require.True(t, ok)
	require.Equal(t, tx.ID(), removedTxn.ID())

	lease.Release()
	time.Sleep(90 * time.Millisecond)
	require.False(t, isTxnCached(cache, tx.ID()))
	require.NoError(t, tx.Commit())
}

func TestTransactionMiddleware_HeldRequestExpiresAfterHandlerReturns(t *testing.T) {
	cdb := setupDatabase(t)
	tx, err := cdb.NewTxn(false)
	require.NoError(t, err)

	cache := testTxnCache(t, 30*time.Millisecond)
	require.NoError(t, cache.Store(tx))

	entered := make(chan struct{})
	releaseHandler := make(chan struct{})
	next := stdHTTP.HandlerFunc(func(rw stdHTTP.ResponseWriter, req *stdHTTP.Request) {
		close(entered)
		<-releaseHandler
		rw.WriteHeader(stdHTTP.StatusOK)
	})
	handler := ApiMiddleware(cdb, cache, nil)(TransactionMiddleware(next))

	req := httptest.NewRequest(stdHTTP.MethodGet, "http://localhost:9181/api/v1/collections/User", nil)
	req.Header.Set(txHeaderName, strconv.FormatUint(tx.ID(), 10))
	rec := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		handler.ServeHTTP(rec, req)
		close(done)
	}()

	<-entered
	time.Sleep(90 * time.Millisecond)
	require.True(t, isTxnCached(cache, tx.ID()))

	close(releaseHandler)
	<-done
	require.Equal(t, stdHTTP.StatusOK, rec.Code)

	require.Eventually(t, func() bool {
		return !isTxnCached(cache, tx.ID())
	}, time.Second, 10*time.Millisecond)
}

// isCached reports whether a transaction with the given ID is currently cached.
func isTxnCached(c *txnCache, id uint64) bool {
	_, ok := c.cache.Load(id)
	return ok
}

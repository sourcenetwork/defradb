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
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	clientmocks "github.com/sourcenetwork/defradb/client/mocks"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/ttl"
)

func TestParseTxnTTL(t *testing.T) {
	t.Run("duration", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "http://localhost:9181/api/v1/tx?ttl=150ms", nil)

		txnTTL, err := parseTxnTTL(req)

		require.NoError(t, err)
		require.Equal(t, 150*time.Millisecond, txnTTL)
	})

	t.Run("seconds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "http://localhost:9181/api/v1/tx?ttl=2", nil)

		txnTTL, err := parseTxnTTL(req)

		require.NoError(t, err)
		require.Equal(t, 2*time.Second, txnTTL)
	})

	t.Run("empty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "http://localhost:9181/api/v1/tx", nil)

		txnTTL, err := parseTxnTTL(req)

		require.NoError(t, err)
		require.Zero(t, txnTTL)
	})

	t.Run("invalid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "http://localhost:9181/api/v1/tx?ttl=soon", nil)

		_, err := parseTxnTTL(req)

		require.Error(t, err)
	})
}

func TestTxHandler_GivenTTL_ExpiresTransaction(t *testing.T) {
	cdb := setupDatabase(t)
	handler, err := NewHandler(cdb, &options.NodeOptions{
		HTTP: options.NodeHTTPOptions{
			TxnTTL:        100 * time.Millisecond,
			TxnTTLTick:    10 * time.Millisecond,
			TxnTTLBuckets: 20,
		},
	})
	require.NoError(t, err)
	defer handler.Close()

	req := httptest.NewRequest(http.MethodPost, "http://localhost:9181/api/v1/tx?ttl=30ms", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Result().StatusCode)
	var response CreateTxResponse
	require.NoError(t, json.NewDecoder(rec.Result().Body).Decode(&response))

	require.Eventually(t, func() bool {
		return !isTxnCached(handler.txs, response.ID)
	}, time.Second, 10*time.Millisecond)

	req = httptest.NewRequest(http.MethodPost, "http://localhost:9181/api/v1/tx/"+strconv.FormatUint(response.ID, 10), nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Result().StatusCode)
}

func TestNewTxnCache_GivenDefaultTTLBeyondMax_ReturnsError(t *testing.T) {
	cache, err := newTxnCache(context.Background(), &options.NodeHTTPOptions{
		TxnTTL:        time.Minute,
		TxnTTLTick:    time.Second,
		TxnTTLBuckets: 1,
	})

	require.ErrorIs(t, err, ttl.ErrBeyondMaxTTL)
	require.Nil(t, cache)
}

func TestNewHandler_GivenNilNodeOptions_UsesDefaultTxnCache(t *testing.T) {
	cdb := setupDatabase(t)

	handler, err := NewHandler(cdb, nil)

	require.NoError(t, err)
	defer handler.Close()
	require.NotNil(t, handler.txs)
	require.Equal(t, DefaultTxnTTL, handler.txs.defaultTTL)
}

func TestTxHandlerCommit_GivenCommitError_RestoresTransaction(t *testing.T) {
	txs, err := newTxnCache(context.Background(), &options.NodeHTTPOptions{
		TxnTTL:        5 * time.Second,
		TxnTTLTick:    time.Second,
		TxnTTLBuckets: 10,
	})
	require.NoError(t, err)
	defer txs.Close()

	const txID uint64 = 1
	txnTTL := 3 * time.Second
	commitErr := errors.New("commit failed")
	tx := clientmocks.NewTxn(t)
	tx.EXPECT().ID().Return(txID)
	tx.EXPECT().Commit().Return(commitErr)
	require.NoError(t, txs.Store(tx, txnTTL))

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", strconv.FormatUint(txID, 10))
	req := httptest.NewRequest(http.MethodPost, "/tx/1", nil)
	ctx := context.WithValue(req.Context(), txsContextKey, txs)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, routeCtx)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	(&txHandler{}).Commit(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
	item, ok := txs.cache.Load(txID)
	require.True(t, ok)
	require.Same(t, tx, item.txn)
	require.Equal(t, txnTTL, item.ttl)
}

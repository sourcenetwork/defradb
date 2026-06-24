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
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

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
		_, ok := handler.txs.cache.Load(response.ID)
		return !ok
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

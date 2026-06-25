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
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

func TestTransactionMiddleware_WithInvalidTxID_ReturnsBadRequest(t *testing.T) {
	cdb := setupDatabase(t)

	handler, err := NewHandler(cdb, nil)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "http://localhost:9181/api/v1/collections?name=User", nil)
	req.Header.Set(txHeaderName, "not-a-number")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	var errRes map[string]any
	require.NoError(t, json.Unmarshal(body, &errRes))
	assert.Contains(t, errRes["error"], ErrInvalidTransactionId.Error())
}

func TestTransactionMiddleware_WithNonExistentTxID_ReturnsNotFound(t *testing.T) {
	cdb := setupDatabase(t)

	handler, err := NewHandler(cdb, nil)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "http://localhost:9181/api/v1/collections?name=User", nil)
	req.Header.Set(txHeaderName, "99999")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	var errRes map[string]any
	require.NoError(t, json.Unmarshal(body, &errRes))
	assert.Contains(t, errRes["error"], client.ErrTransactionNotFound.Error())
}

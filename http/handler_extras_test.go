// Copyright 2024 Democratized Data Foundation
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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/event"

	"github.com/stretchr/testify/require"
)

func TestPurgeDevModeTrue(t *testing.T) {
	cdb := setupDatabase(t)

	url := "http://localhost:9181/api/purge"

	req := httptest.NewRequest(http.MethodPost, url, nil)
	rec := httptest.NewRecorder()

	purgeSub, err := cdb.Events().Subscribe(event.PurgeName)
	require.NoError(t, err)

	handler, err := NewHandler(cdb, &options.NodeOptions{EnableDevelopment: true})
	require.NoError(t, err)
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	require.Equal(t, 200, res.StatusCode)

	// test will timeout if purge never received
	<-purgeSub.Message()
}

func TestPurgeDevModeFalse(t *testing.T) {
	cdb := setupDatabase(t)

	url := "http://localhost:9181/api/purge"

	req := httptest.NewRequest(http.MethodPost, url, nil)
	rec := httptest.NewRecorder()

	handler, err := NewHandler(cdb, nil)
	require.NoError(t, err)
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	require.Equal(t, 403, res.StatusCode)
}

// Copyright 2023 Democratized Data Foundation
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
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore/memory"
	"github.com/sourcenetwork/defradb/internal/db"
)

func TestCCIPGet_WithValidData(t *testing.T) {
	cdb := setupDatabase(t)

	gqlData, err := json.Marshal(&GraphQLRequest{
		Query: `query {
			User {
				name
			}
		}`,
	})
	require.NoError(t, err)

	data := "0x" + hex.EncodeToString([]byte(gqlData))
	sender := "0x0000000000000000000000000000000000000000"
	url := "http://localhost:9181/api/v0/ccip/" + path.Join(sender, data)

	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()

	handler, err := NewHandler(cdb)
	require.NoError(t, err)
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	require.NotNil(t, res.Body)

	resData, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	var ccipRes CCIPResponse
	err = json.Unmarshal(resData, &ccipRes)
	require.NoError(t, err)

	resHex, err := hex.DecodeString(strings.TrimPrefix(ccipRes.Data, "0x"))
	require.NoError(t, err)

	assert.JSONEq(t, `{"data": [{"name": "bob"}]}`, string(resHex))
}

func TestCCIPGet_WithSubscription(t *testing.T) {
	cdb := setupDatabase(t)

	gqlData, err := json.Marshal(&GraphQLRequest{
		Query: `subscription {
			User {
				name
			}
		}`,
	})
	require.NoError(t, err)

	data := "0x" + hex.EncodeToString([]byte(gqlData))
	sender := "0x0000000000000000000000000000000000000000"
	url := "http://localhost:9181/api/v0/ccip/" + path.Join(sender, data)

	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()

	handler, err := NewHandler(cdb)
	require.NoError(t, err)
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	assert.Equal(t, 400, res.StatusCode)
}

func TestCCIPGet_WithInvalidData(t *testing.T) {
	cdb := setupDatabase(t)

	data := "invalid_hex_data"
	sender := "0x0000000000000000000000000000000000000000"
	url := "http://localhost:9181/api/v0/ccip/" + path.Join(sender, data)

	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()

	handler, err := NewHandler(cdb)
	require.NoError(t, err)
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	assert.Equal(t, 400, res.StatusCode)
}

func TestCCIPPost_WithValidData(t *testing.T) {
	cdb := setupDatabase(t)

	gqlJSON, err := json.Marshal(&GraphQLRequest{
		Query: `query {
			User {
				name
			}
		}`,
	})
	require.NoError(t, err)

	body, err := json.Marshal(&CCIPRequest{
		Data:   "0x" + hex.EncodeToString([]byte(gqlJSON)),
		Sender: "0x0000000000000000000000000000000000000000",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "http://localhost:9181/api/v0/ccip", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler, err := NewHandler(cdb)
	require.NoError(t, err)
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	require.NotNil(t, res.Body)

	resData, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	var ccipRes CCIPResponse
	err = json.Unmarshal(resData, &ccipRes)
	require.NoError(t, err)

	resHex, err := hex.DecodeString(strings.TrimPrefix(ccipRes.Data, "0x"))
	require.NoError(t, err)

	assert.JSONEq(t, `{"data": [{"name": "bob"}]}`, string(resHex))
}

func TestCCIPPost_WithInvalidGraphQLRequest(t *testing.T) {
	cdb := setupDatabase(t)

	body, err := json.Marshal(&CCIPRequest{
		Data:   "0x" + hex.EncodeToString([]byte("invalid_graphql_request")),
		Sender: "0x0000000000000000000000000000000000000000",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "http://localhost:9181/api/v0/ccip", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler, err := NewHandler(cdb)
	require.NoError(t, err)
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	assert.Equal(t, 400, res.StatusCode)
}

func TestCCIPPost_WithInvalidBody(t *testing.T) {
	cdb := setupDatabase(t)

	req := httptest.NewRequest(http.MethodPost, "http://localhost:9181/api/v0/ccip", nil)
	rec := httptest.NewRecorder()

	handler, err := NewHandler(cdb)
	require.NoError(t, err)
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	assert.Equal(t, 400, res.StatusCode)
}

func setupDatabase(t *testing.T) client.DB {
	ctx := context.Background()

	cdb, err := db.NewDB(ctx, memory.NewDatastore(ctx), acp.NoACP, db.WithUpdateEvents())
	require.NoError(t, err)

	_, err = cdb.AddSchema(ctx, `type User {
		name: String
	}`)
	require.NoError(t, err)

	col, err := cdb.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "bob"}`), col.Definition())
	require.NoError(t, err)

	err = col.Create(ctx, doc)
	require.NoError(t, err)

	return cdb
}

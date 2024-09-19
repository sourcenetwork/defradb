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
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecRequest_WithValidQuery_OmitsErrors(t *testing.T) {
	cdb := setupDatabase(t)

	body, err := json.Marshal(&GraphQLRequest{
		Query: `query {
			User {
				name
			}
		}`,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "http://localhost:9181/api/v0/graphql", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler, err := NewHandler(cdb)
	require.NoError(t, err)
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	require.NotNil(t, res.Body)

	resData, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	var gqlResponse map[string]any
	err = json.Unmarshal(resData, &gqlResponse)
	require.NoError(t, err)

	// errors should be omitted
	_, ok := gqlResponse["errors"]
	assert.False(t, ok)
}

func TestExecRequest_WithInvalidQuery_HasSpecCompliantErrors(t *testing.T) {
	cdb := setupDatabase(t)

	body, err := json.Marshal(&GraphQLRequest{
		Query: `query {
			User {
				invalid
			}
		}`,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "http://localhost:9181/api/v0/graphql", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler, err := NewHandler(cdb)
	require.NoError(t, err)
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	require.NotNil(t, res.Body)

	resData, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	var gqlResponse map[string]any
	err = json.Unmarshal(resData, &gqlResponse)
	require.NoError(t, err)

	errList, ok := gqlResponse["errors"]
	require.True(t, ok)

	// errors should contain spec compliant error objects
	assert.ElementsMatch(t, errList, []any{map[string]any{
		"message": "Cannot query field \"invalid\" on type \"User\".",
	}})
}

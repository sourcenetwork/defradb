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
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// chi treats `collections` and `collections/` as separate routes, so without
// normalization the trailing-slash form 404s. Verify StripSlashes routes both to the
// same handler.
func TestHandler_GetCollectionWithTrailingSlash(t *testing.T) {
	cdb := setupDatabase(t)

	handler, err := NewHandler(cdb, nil)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "http://localhost:9181/api/v1/collections/?name=User", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equalf(t, http.StatusOK, res.StatusCode, "expected 200, got %d: %s", res.StatusCode, string(body))
}

// Trailing-slash normalization for the collection delete endpoint; see the GET case.
func TestHandler_DeleteCollectionWithTrailingSlash(t *testing.T) {
	cdb := setupDatabase(t)

	// delete refuses a collection with documents, so use an empty one
	_, err := cdb.AddCollection(context.Background(), `type Author {
		name: String
	}`)
	require.NoError(t, err)

	handler, err := NewHandler(cdb, nil)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "http://localhost:9181/api/v1/collections/?name=Author", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equalf(t, http.StatusOK, res.StatusCode, "expected 200, got %d: %s", res.StatusCode, string(body))
}

// The GET /collections response renders each field's Kind and Typ as human-readable
// strings (issue #4816), not their opaque numeric IDs. Assert the string form appears on
// the wire and the numeric form does not. The round-trip back into structs is covered by
// the integration suite running under DEFRA_CLIENT_TYPE=http.
func TestHandler_GetCollection_RendersStringKindAndTyp(t *testing.T) {
	cdb := setupDatabase(t)

	handler, err := NewHandler(cdb, nil)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "http://localhost:9181/api/v1/collections?name=User", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equalf(t, http.StatusOK, res.StatusCode, "expected 200, got %d: %s", res.StatusCode, string(body))

	out := string(body)
	// The `name` field is a nillable String with an LWW CRDT.
	assert.Contains(t, out, `"Kind":"String"`)
	assert.Contains(t, out, `"Typ":"lww"`)
	// The numeric forms must not leak into the display output.
	assert.NotContains(t, out, `"Kind":11`)
	assert.NotContains(t, out, `"Typ":1`)
}

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

	req := httptest.NewRequest(http.MethodPost, "http://localhost:9181/api/graphql", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler, err := NewHandler(cdb, nil)
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

	req := httptest.NewRequest(http.MethodPost, "http://localhost:9181/api/graphql", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler, err := NewHandler(cdb, nil)
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

func TestExecRequest_HttpGet_WithOperationName(t *testing.T) {
	cdb := setupDatabase(t)

	query := `
	query UserQuery {
		User {
			name
		}
	}
	query UserQueryWithDocID {
		User {
			_docID
			name
		}
	}
	`
	operationName := "UserQuery"

	encodedQuery := url.QueryEscape(query)
	encodedOperationName := url.QueryEscape(operationName)

	endpointURL := "http://localhost:9181/api/graphql?query=" + encodedQuery + "&operationName=" + encodedOperationName

	req := httptest.NewRequest(http.MethodGet, endpointURL, nil)
	rec := httptest.NewRecorder()

	handler, err := NewHandler(cdb, nil)
	require.NoError(t, err)
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	require.NotNil(t, res.Body)

	resData, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	var gqlResponse map[string]any
	err = json.Unmarshal(resData, &gqlResponse)
	require.NoError(t, err)

	// Ensure the response data contains names, but not the _docID field
	expectedJSON := `{
		"data": {
			"User": [
				{"name": "bob"},
				{"name": "adam"}
			]
		}
	}`
	assert.JSONEq(t, expectedJSON, string(resData))
}

func TestExecRequest_HttpGet_WithVariables(t *testing.T) {
	cdb := setupDatabase(t)

	query := `query getUser($filter: UserFilterArg) {
		User(filter: $filter) {
			name
		}
	}`
	operationName := "getUser"
	variables := `{"filter":{"name":{"_eq":"bob"}}}`

	encodedQuery := url.QueryEscape(query)
	encodedOperationName := url.QueryEscape(operationName)
	encodedVariables := url.QueryEscape(variables)

	endpointURL := "http://localhost:9181/api/graphql?query=" + encodedQuery + "&operationName=" + encodedOperationName + "&variables=" + encodedVariables

	req := httptest.NewRequest(http.MethodGet, endpointURL, nil)
	rec := httptest.NewRecorder()

	handler, err := NewHandler(cdb, nil)
	require.NoError(t, err)
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	require.NotNil(t, res.Body)

	resData, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	var gqlResponse map[string]any
	err = json.Unmarshal(resData, &gqlResponse)
	require.NoError(t, err)

	// Ensure only bob is returned, because of the filter variable
	expectedJSON := `{
		"data": {
			"User": [
				{"name": "bob"}
			]
		}
	}`
	assert.JSONEq(t, expectedJSON, string(resData))
}

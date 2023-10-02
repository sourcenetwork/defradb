// Copyright 2022 Democratized Data Foundation
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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	dshelp "github.com/ipfs/boxo/datastore/dshelp"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v4"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/errors"
)

type testOptions struct {
	Testing        *testing.T
	DB             client.DB
	Handlerfunc    http.HandlerFunc
	Method         string
	Path           string
	Body           io.Reader
	Headers        map[string]string
	QueryParams    map[string]string
	ExpectedStatus int
	ResponseData   any
	ServerOptions  serverOptions
}

type testUser struct {
	Key      string        `json:"_key"`
	Versions []testVersion `json:"_version"`
}

type testVersion struct {
	CID string `json:"cid"`
}

func TestRootHandler(t *testing.T) {
	resp := DataResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             nil,
		Method:         "GET",
		Path:           RootPath,
		Body:           nil,
		ExpectedStatus: 200,
		ResponseData:   &resp,
	})
	switch v := resp.Data.(type) {
	case map[string]any:
		require.Equal(t, "Welcome to the DefraDB HTTP API. Use /graphql to send queries to the database. Read the documentation at https://docs.source.network/.", v["response"])
	default:
		t.Fatalf("data should be of type map[string]any but got %T", resp.Data)
	}
}

func TestPingHandler(t *testing.T) {
	resp := DataResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             nil,
		Method:         "GET",
		Path:           PingPath,
		Body:           nil,
		ExpectedStatus: 200,
		ResponseData:   &resp,
	})

	switch v := resp.Data.(type) {
	case map[string]any:
		require.Equal(t, "pong", v["response"])
	default:
		t.Fatalf("data should be of type map[string]any but got %T", resp.Data)
	}
}

func TestDumpHandlerWithNoError(t *testing.T) {
	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	resp := DataResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "GET",
		Path:           DumpPath,
		Body:           nil,
		ExpectedStatus: 200,
		ResponseData:   &resp,
	})

	switch v := resp.Data.(type) {
	case map[string]any:
		require.Equal(t, "ok", v["response"])
	default:
		t.Fatalf("data should be of type map[string]any but got %T", resp.Data)
	}
}

func TestDumpHandlerWithDBError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             nil,
		Method:         "GET",
		Path:           DumpPath,
		Body:           nil,
		ExpectedStatus: 500,
		ResponseData:   &errResponse,
	})
	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "no database available")
	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "no database available", errResponse.Errors[0].Message)
}

func TestExecGQLWithNilBody(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             nil,
		Method:         "POST",
		Path:           GraphQLPath,
		Body:           nil,
		ExpectedStatus: 400,
		ResponseData:   &errResponse,
	})

	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "body cannot be empty")
	require.Equal(t, http.StatusBadRequest, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Bad Request", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "body cannot be empty", errResponse.Errors[0].Message)
}

func TestExecGQLWithEmptyBody(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             nil,
		Method:         "POST",
		Path:           GraphQLPath,
		Body:           bytes.NewBuffer([]byte("")),
		ExpectedStatus: 400,
		ResponseData:   &errResponse,
	})

	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "missing GraphQL request")
	require.Equal(t, http.StatusBadRequest, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Bad Request", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "missing GraphQL request", errResponse.Errors[0].Message)
}

type mockReadCloser struct {
	mock.Mock
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func TestExecGQLWithMockBody(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	mockReadCloser := mockReadCloser{}
	// if Read is called, it will return error
	mockReadCloser.On("Read", mock.AnythingOfType("[]uint8")).Return(0, errors.New("error reading"))

	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             nil,
		Method:         "POST",
		Path:           GraphQLPath,
		Body:           &mockReadCloser,
		ExpectedStatus: 500,
		ResponseData:   &errResponse,
	})

	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "error reading")
	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "error reading", errResponse.Errors[0].Message)
}

func TestExecGQLWithInvalidContentType(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	errResponse := ErrorResponse{}
	stmt := `
mutation {
	create_User(data: "{\"age\": 31, \"verified\": true, \"points\": 90, \"name\": \"Bob\"}") {
		_key
	}
}`

	buf := bytes.NewBuffer([]byte(stmt))
	testRequest(testOptions{
		Testing:        t,
		Method:         "POST",
		Path:           GraphQLPath,
		Body:           buf,
		ExpectedStatus: 500,
		Headers:        map[string]string{"Content-Type": contentTypeJSON + "; this-is-wrong"},
		ResponseData:   &errResponse,
	})

	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "mime: invalid media parameter")
	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "mime: invalid media parameter", errResponse.Errors[0].Message)
}

func TestExecGQLWithNoDB(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	errResponse := ErrorResponse{}
	stmt := `
mutation {
	create_User(data: "{\"age\": 31, \"verified\": true, \"points\": 90, \"name\": \"Bob\"}") {
		_key
	}
}`

	buf := bytes.NewBuffer([]byte(stmt))
	testRequest(testOptions{
		Testing:        t,
		Method:         "POST",
		Path:           GraphQLPath,
		Body:           buf,
		ExpectedStatus: 500,
		ResponseData:   &errResponse,
	})

	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "no database available")
	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "no database available", errResponse.Errors[0].Message)
}

func TestExecGQLHandlerContentTypeJSONWithJSONError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	// statement with JSON formatting error
	stmt := `
[
	"query": "mutation {
		create_User(
			data: \"{
				\\\"age\\\": 31,
				\\\"verified\\\": true,
				\\\"points\\\": 90,
				\\\"name\\\": \\\"Bob\\\"
			}\"
		) {_key}
	}"
]`

	buf := bytes.NewBuffer([]byte(stmt))
	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             nil,
		Method:         "POST",
		Path:           GraphQLPath,
		Body:           buf,
		Headers:        map[string]string{"Content-Type": contentTypeJSON},
		ExpectedStatus: 400,
		ResponseData:   &errResponse,
	})

	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "invalid character")
	require.Equal(t, http.StatusBadRequest, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Bad Request", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "unmarshal error: invalid character ':' after array element", errResponse.Errors[0].Message)
}

func TestExecGQLHandlerContentTypeJSON(t *testing.T) {
	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	// load schema
	testLoadSchema(t, ctx, defra)

	// add document
	stmt := `
{
	"query": "mutation {
		create_User(
			data: \"{
				\\\"age\\\": 31,
				\\\"verified\\\": true,
				\\\"points\\\": 90,
				\\\"name\\\": \\\"Bob\\\"
			}\"
		) {_key}
	}"
}`
	// remove line returns and tabulation from formatted statement
	stmt = strings.ReplaceAll(strings.ReplaceAll(stmt, "\t", ""), "\n", "")

	buf := bytes.NewBuffer([]byte(stmt))
	users := []testUser{}
	resp := DataResponse{
		Data: &users,
	}
	testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           GraphQLPath,
		Body:           buf,
		Headers:        map[string]string{"Content-Type": contentTypeJSON},
		ExpectedStatus: 200,
		ResponseData:   &resp,
	})

	require.Contains(t, users[0].Key, "bae-")
}

func TestExecGQLHandlerContentTypeJSONWithError(t *testing.T) {
	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	// load schema
	testLoadSchema(t, ctx, defra)

	// add document
	stmt := `
	{
		"query": "mutation {
			create_User(
				data: \"{
					\\\"age\\\": 31,
					\\\"notAField\\\": true
				}\"
			) {_key}
		}"
	}`

	// remove line returns and tabulation from formatted statement
	stmt = strings.ReplaceAll(strings.ReplaceAll(stmt, "\t", ""), "\n", "")

	buf := bytes.NewBuffer([]byte(stmt))
	resp := GQLResult{}
	testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           GraphQLPath,
		Body:           buf,
		Headers:        map[string]string{"Content-Type": contentTypeJSON},
		ExpectedStatus: 200,
		ResponseData:   &resp,
	})

	require.Contains(t, resp.Errors, "The given field does not exist. Name: notAField")
	require.Len(t, resp.Errors, 1)
}

func TestExecGQLHandlerContentTypeJSONWithCharset(t *testing.T) {
	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	// load schema
	testLoadSchema(t, ctx, defra)

	// add document
	stmt := `
{
	"query": "mutation {
		create_User(
			data: \"{
				\\\"age\\\": 31,
				\\\"verified\\\": true,
				\\\"points\\\": 90,
				\\\"name\\\": \\\"Bob\\\"
			}\"
		) {_key}
	}"
}`
	// remote line returns and tabulation from formatted statement
	stmt = strings.ReplaceAll(strings.ReplaceAll(stmt, "\t", ""), "\n", "")

	buf := bytes.NewBuffer([]byte(stmt))
	users := []testUser{}
	resp := DataResponse{
		Data: &users,
	}
	testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           GraphQLPath,
		Body:           buf,
		Headers:        map[string]string{"Content-Type": contentTypeJSON + "; charset=utf8"},
		ExpectedStatus: 200,
		ResponseData:   &resp,
	})

	require.Contains(t, users[0].Key, "bae-")
}

func TestExecGQLHandlerContentTypeFormURLEncoded(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             nil,
		Method:         "POST",
		Path:           GraphQLPath,
		Body:           nil,
		Headers:        map[string]string{"Content-Type": contentTypeFormURLEncoded},
		ExpectedStatus: 400,
		ResponseData:   &errResponse,
	})

	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "content type application/x-www-form-urlencoded not yet supported")
	require.Equal(t, http.StatusBadRequest, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Bad Request", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "content type application/x-www-form-urlencoded not yet supported", errResponse.Errors[0].Message)
}

func TestExecGQLHandlerContentTypeGraphQL(t *testing.T) {
	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	// load schema
	testLoadSchema(t, ctx, defra)

	// add document
	stmt := `
mutation {
	create_User(data: "{\"age\": 31, \"verified\": true, \"points\": 90, \"name\": \"Bob\"}") {
		_key
	}
}`

	buf := bytes.NewBuffer([]byte(stmt))
	users := []testUser{}
	resp := DataResponse{
		Data: &users,
	}
	testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           GraphQLPath,
		Body:           buf,
		Headers:        map[string]string{"Content-Type": contentTypeGraphQL},
		ExpectedStatus: 200,
		ResponseData:   &resp,
	})

	require.Contains(t, users[0].Key, "bae-")
}

func TestExecGQLHandlerContentTypeText(t *testing.T) {
	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	// load schema
	testLoadSchema(t, ctx, defra)

	// add document
	stmt := `
mutation {
	create_User(data: "{\"age\": 31, \"verified\": true, \"points\": 90, \"name\": \"Bob\"}") {
		_key
	}
}`

	buf := bytes.NewBuffer([]byte(stmt))
	users := []testUser{}
	resp := DataResponse{
		Data: &users,
	}
	testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           GraphQLPath,
		Body:           buf,
		ExpectedStatus: 200,
		ResponseData:   &resp,
	})

	require.Contains(t, users[0].Key, "bae-")
}

func TestExecGQLHandlerWithSubsctiption(t *testing.T) {
	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	// load schema
	testLoadSchema(t, ctx, defra)

	stmt := `
subscription {
	User {
		_key
		age
		name
	}
}`

	buf := bytes.NewBuffer([]byte(stmt))

	ch := make(chan []byte)
	errCh := make(chan error)

	// We need to set a timeout otherwise the testSubscriptionRequest function will block until the
	// http.ServeHTTP call returns, which in this case will only happen with a timeout.
	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	go testSubscriptionRequest(ctxTimeout, testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           GraphQLPath,
		Body:           buf,
		Headers:        map[string]string{"Content-Type": contentTypeGraphQL},
		ExpectedStatus: 200,
	}, ch, errCh)

	// We wait to ensure the subscription requests can subscribe to the event channel.
	time.Sleep(time.Second / 2)

	// add document
	stmt2 := `
mutation {
	create_User(data: "{\"age\": 31, \"verified\": true, \"points\": 90, \"name\": \"Bob\"}") {
		_key
	}
}`

	buf2 := bytes.NewBuffer([]byte(stmt2))
	users := []testUser{}
	resp := DataResponse{
		Data: &users,
	}
	testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           GraphQLPath,
		Body:           buf2,
		ExpectedStatus: 200,
		ResponseData:   &resp,
	})
	select {
	case data := <-ch:
		require.Contains(t, string(data), users[0].Key)
	case err := <-errCh:
		t.Fatal(err)
	}
}

func TestListSchemaHandlerWithoutDB(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"

	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             nil,
		Method:         "GET",
		Path:           SchemaPath,
		ExpectedStatus: 500,
		ResponseData:   &errResponse,
	})

	assert.Contains(t, errResponse.Errors[0].Extensions.Stack, "no database available")
	assert.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
	assert.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
	assert.Equal(t, "no database available", errResponse.Errors[0].Message)
}

func TestListSchemaHandlerWitNoError(t *testing.T) {
	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	stmt := `
type user {
	name: String
	age: Int
	verified: Boolean
	points: Float
}
type group {
	owner: user
	members: [user]
}`

	_, err := defra.AddSchema(ctx, stmt)
	if err != nil {
		t.Fatal(err)
	}

	resp := DataResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "GET",
		Path:           SchemaPath,
		ExpectedStatus: 200,
		ResponseData:   &resp,
	})

	switch v := resp.Data.(type) {
	case map[string]any:
		assert.Equal(t, map[string]any{
			"collections": []any{
				map[string]any{
					"name":       "group",
					"id":         "bafkreicdtcgmgjjjao4zzaoacy26cl7xtnnev4qotvflellmdrzi57m5re",
					"version_id": "bafkreicdtcgmgjjjao4zzaoacy26cl7xtnnev4qotvflellmdrzi57m5re",
					"fields": []any{
						map[string]any{
							"id":       "0",
							"kind":     "ID",
							"name":     "_key",
							"internal": true,
						},
						map[string]any{
							"id":       "1",
							"kind":     "[user]",
							"name":     "members",
							"internal": false,
						},
						map[string]any{
							"id":       "2",
							"kind":     "user",
							"name":     "owner",
							"internal": false,
						},
						map[string]any{
							"id":       "3",
							"kind":     "ID",
							"name":     "owner_id",
							"internal": true,
						},
					},
				},
				map[string]any{
					"name":       "user",
					"id":         "bafkreigl2v5trzfznb7dm3dubmsbzkw73s3phjm6laegswzl4625wc2grm",
					"version_id": "bafkreigl2v5trzfznb7dm3dubmsbzkw73s3phjm6laegswzl4625wc2grm",
					"fields": []any{
						map[string]any{
							"id":       "0",
							"kind":     "ID",
							"name":     "_key",
							"internal": true,
						},
						map[string]any{
							"id":       "1",
							"kind":     "Int",
							"name":     "age",
							"internal": false,
						},
						map[string]any{
							"id":       "2",
							"kind":     "String",
							"name":     "name",
							"internal": false,
						},
						map[string]any{
							"id":       "3",
							"kind":     "Float",
							"name":     "points",
							"internal": false,
						},
						map[string]any{
							"id":       "4",
							"kind":     "Boolean",
							"name":     "verified",
							"internal": false,
						},
					},
				},
			},
		}, v)

	default:
		t.Fatalf("data should be of type map[string]any but got %T\n%v", resp.Data, v)
	}
}

func TestLoadSchemaHandlerWithReadBodyError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	mockReadCloser := mockReadCloser{}
	// if Read is called, it will return error
	mockReadCloser.On("Read", mock.AnythingOfType("[]uint8")).Return(0, errors.New("error reading"))

	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             nil,
		Method:         "POST",
		Path:           SchemaPath,
		Body:           &mockReadCloser,
		ExpectedStatus: 500,
		ResponseData:   &errResponse,
	})

	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "error reading")
	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "error reading", errResponse.Errors[0].Message)
}

func TestLoadSchemaHandlerWithoutDB(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	stmt := `
type User {
	name: String
	age: Int
	verified: Boolean
	points: Float
}`

	buf := bytes.NewBuffer([]byte(stmt))

	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             nil,
		Method:         "POST",
		Path:           SchemaPath,
		Body:           buf,
		ExpectedStatus: 500,
		ResponseData:   &errResponse,
	})

	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "no database available")
	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "no database available", errResponse.Errors[0].Message)
}

func TestLoadSchemaHandlerWithAddSchemaError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	// statement with types instead of type
	stmt := `
types User {
	name: String
	age: Int
	verified: Boolean
	points: Float
}`

	buf := bytes.NewBuffer([]byte(stmt))

	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           SchemaPath,
		Body:           buf,
		ExpectedStatus: 500,
		ResponseData:   &errResponse,
	})

	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "Syntax Error GraphQL (2:1) Unexpected Name")
	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(
		t,
		"Syntax Error GraphQL (2:1) Unexpected Name \"types\"\n\n1: \n2: types User {\n   ^\n3: \\u0009name: String\n",
		errResponse.Errors[0].Message,
	)
}

func TestLoadSchemaHandlerWitNoError(t *testing.T) {
	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	stmt := `
type User {
	name: String
	age: Int
	verified: Boolean
	points: Float
}`

	buf := bytes.NewBuffer([]byte(stmt))

	resp := DataResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           SchemaPath,
		Body:           buf,
		ExpectedStatus: 200,
		ResponseData:   &resp,
	})

	switch v := resp.Data.(type) {
	case map[string]any:
		require.Equal(t, map[string]any{
			"result": "success",
			"collections": []any{
				map[string]any{
					"name":       "User",
					"id":         "bafkreiet7xqehjsjsthy6nafvtbz4el376uudhkjyeifuvvsr64se33swm",
					"version_id": "bafkreiet7xqehjsjsthy6nafvtbz4el376uudhkjyeifuvvsr64se33swm",
				},
			},
		}, v)

	default:
		t.Fatalf("data should be of type map[string]any but got %T\n%v", resp.Data, v)
	}
}

func TestGetBlockHandlerWithMultihashError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             nil,
		Method:         "GET",
		Path:           BlocksPath + "/1234",
		Body:           nil,
		ExpectedStatus: 400,
		ResponseData:   &errResponse,
	})

	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "illegal base32 data at input byte 0")
	require.Equal(t, http.StatusBadRequest, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Bad Request", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "illegal base32 data at input byte 0", errResponse.Errors[0].Message)
}

func TestGetBlockHandlerWithDSKeyWithNoDB(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	cID, err := cid.Parse("bafybeidembipteezluioakc2zyke4h5fnj4rr3uaougfyxd35u3qzefzhm")
	if err != nil {
		t.Fatal(err)
	}
	dsKey := dshelp.MultihashToDsKey(cID.Hash())

	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             nil,
		Method:         "GET",
		Path:           BlocksPath + dsKey.String(),
		Body:           nil,
		ExpectedStatus: 500,
		ResponseData:   &errResponse,
	})

	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "no database available")
	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "no database available", errResponse.Errors[0].Message)
}

func TestGetBlockHandlerWithNoDB(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             nil,
		Method:         "GET",
		Path:           BlocksPath + "/bafybeidembipteezluioakc2zyke4h5fnj4rr3uaougfyxd35u3qzefzhm",
		Body:           nil,
		ExpectedStatus: 500,
		ResponseData:   &errResponse,
	})

	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "no database available")
	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "no database available", errResponse.Errors[0].Message)
}

func TestGetBlockHandlerWithGetBlockstoreError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "GET",
		Path:           BlocksPath + "/bafybeidembipteezluioakc2zyke4h5fnj4rr3uaougfyxd35u3qzefzhm",
		Body:           nil,
		ExpectedStatus: 500,
		ResponseData:   &errResponse,
	})

	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "ipld: could not find bafybeidembipteezluioakc2zyke4h5fnj4rr3uaougfyxd35u3qzefzhm")
	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "ipld: could not find bafybeidembipteezluioakc2zyke4h5fnj4rr3uaougfyxd35u3qzefzhm", errResponse.Errors[0].Message)
}

func TestGetBlockHandlerWithValidBlockstore(t *testing.T) {
	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	testLoadSchema(t, ctx, defra)

	// add document
	stmt := `
mutation {
	create_User(data: "{\"age\": 31, \"verified\": true, \"points\": 90, \"name\": \"Bob\"}") {
		_key
	}
}`

	buf := bytes.NewBuffer([]byte(stmt))

	users := []testUser{}
	resp := DataResponse{
		Data: &users,
	}
	testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           GraphQLPath,
		Body:           buf,
		ExpectedStatus: 200,
		ResponseData:   &resp,
	})

	if !strings.Contains(users[0].Key, "bae-") {
		t.Fatal("expected valid document key")
	}

	// get document cid
	stmt2 := `
query {
	User (dockey: "%s") {
		_version {
			cid
		}
	}
}`
	buf2 := bytes.NewBuffer([]byte(fmt.Sprintf(stmt2, users[0].Key)))

	users2 := []testUser{}
	resp2 := DataResponse{
		Data: &users2,
	}
	testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           GraphQLPath,
		Body:           buf2,
		ExpectedStatus: 200,
		ResponseData:   &resp2,
	})

	_, err := cid.Decode(users2[0].Versions[0].CID)
	if err != nil {
		t.Fatal(err)
	}

	resp3 := DataResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "GET",
		Path:           BlocksPath + "/" + users2[0].Versions[0].CID,
		Body:           buf,
		ExpectedStatus: 200,
		ResponseData:   &resp3,
	})

	switch d := resp3.Data.(type) {
	case map[string]any:
		switch val := d["val"].(type) {
		case string:
			require.Equal(t, "pGNhZ2UYH2RuYW1lY0JvYmZwb2ludHMYWmh2ZXJpZmllZPU=", val)
		default:
			t.Fatalf("expecting string but got %T", val)
		}
	default:
		t.Fatalf("expecting map[string]any but got %T", d)
	}
}

func TestPeerIDHandler(t *testing.T) {
	resp := DataResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             nil,
		Method:         "GET",
		Path:           PeerIDPath,
		Body:           nil,
		ExpectedStatus: 200,
		ResponseData:   &resp,
		ServerOptions: serverOptions{
			peerID: "12D3KooWFpi6VTYKLtxUftJKEyfX8jDfKi8n15eaygH8ggfYFZbR",
		},
	})

	switch v := resp.Data.(type) {
	case map[string]any:
		require.Equal(t, "12D3KooWFpi6VTYKLtxUftJKEyfX8jDfKi8n15eaygH8ggfYFZbR", v["peerID"])
	default:
		t.Fatalf("data should be of type map[string]any but got %T", resp.Data)
	}
}

func TestPeerIDHandlerWithNoPeerIDInContext(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"

	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             nil,
		Method:         "GET",
		Path:           PeerIDPath,
		Body:           nil,
		ExpectedStatus: 404,
		ResponseData:   &errResponse,
	})

	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "no PeerID available. P2P might be disabled")
	require.Equal(t, http.StatusNotFound, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Not Found", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "no PeerID available. P2P might be disabled", errResponse.Errors[0].Message)
}

func testRequest(opt testOptions) []byte {
	req, err := http.NewRequest(opt.Method, opt.Path, opt.Body)
	if err != nil {
		opt.Testing.Fatal(err)
	}

	for k, v := range opt.Headers {
		req.Header.Set(k, v)
	}

	q := req.URL.Query()
	for k, v := range opt.QueryParams {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	h := newHandler(opt.DB, opt.ServerOptions)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(opt.Testing, opt.ExpectedStatus, rec.Result().StatusCode)

	resBody, err := io.ReadAll(rec.Result().Body)
	if err != nil {
		opt.Testing.Fatal(err)
	}

	if opt.ResponseData != nil {
		err = json.Unmarshal(resBody, &opt.ResponseData)
		if err != nil {
			opt.Testing.Fatal(err)
		}
	}

	return resBody
}

func testSubscriptionRequest(ctx context.Context, opt testOptions, ch chan []byte, errCh chan error) {
	req, err := http.NewRequest(opt.Method, opt.Path, opt.Body)
	if err != nil {
		errCh <- err
		return
	}

	req = req.WithContext(ctx)

	for k, v := range opt.Headers {
		req.Header.Set(k, v)
	}

	h := newHandler(opt.DB, opt.ServerOptions)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(opt.Testing, opt.ExpectedStatus, rec.Result().StatusCode)

	respBody, err := io.ReadAll(rec.Result().Body)
	if err != nil {
		errCh <- err
		return
	}

	ch <- respBody
}

func testNewInMemoryDB(t *testing.T, ctx context.Context) client.DB {
	// init in memory DB
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	if err != nil {
		t.Fatal(err)
	}

	options := []db.Option{
		db.WithUpdateEvents(),
	}

	defra, err := db.NewDB(ctx, rootstore, options...)
	if err != nil {
		t.Fatal(err)
	}

	return defra
}

func testLoadSchema(t *testing.T, ctx context.Context, db client.DB) {
	stmt := `
type User {
	name: String 
	age: Int 
	verified: Boolean 
	points: Float
}`
	_, err := db.AddSchema(ctx, stmt)
	if err != nil {
		t.Fatal(err)
	}
}

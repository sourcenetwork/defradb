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
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/mocks"
	"github.com/sourcenetwork/defradb/errors"
)

func TestExportHandler_WithNoDB_NoDatabaseAvailableError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             nil,
		Method:         "POST",
		Path:           ExportPath,
		Body:           nil,
		ExpectedStatus: 500,
		ResponseData:   &errResponse,
	})
	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "no database available")
	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "no database available", errResponse.Errors[0].Message)
}

func TestExportHandler_WithWrongPayload_ReturnError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	buf := bytes.NewBuffer([]byte("[]"))
	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           ExportPath,
		Body:           buf,
		ExpectedStatus: 400,
		ResponseData:   &errResponse,
	})
	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "json: cannot unmarshal array into Go value of type client.BackupConfig")
	require.Equal(t, http.StatusBadRequest, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Bad Request", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "unmarshal error: json: cannot unmarshal array into Go value of type client.BackupConfig", errResponse.Errors[0].Message)
}

func TestExportHandler_WithInvalidFilePath_ReturnError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	filepath := t.TempDir() + "/some/test.json"
	cfg := client.BackupConfig{
		Filepath: filepath,
	}
	b, err := json.Marshal(cfg)
	require.NoError(t, err)
	buf := bytes.NewBuffer(b)

	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           ExportPath,
		Body:           buf,
		ExpectedStatus: 400,
		ResponseData:   &errResponse,
	})
	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "invalid file path")
	require.Equal(t, http.StatusBadRequest, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Bad Request", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "invalid file path", errResponse.Errors[0].Message)
}

func TestExportHandler_WithInvalidFomat_ReturnError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	filepath := t.TempDir() + "/test.json"
	cfg := client.BackupConfig{
		Filepath: filepath,
		Format:   "csv",
	}
	b, err := json.Marshal(cfg)
	require.NoError(t, err)
	buf := bytes.NewBuffer(b)

	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           ExportPath,
		Body:           buf,
		ExpectedStatus: 400,
		ResponseData:   &errResponse,
	})
	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "only JSON format is supported at the moment")
	require.Equal(t, http.StatusBadRequest, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Bad Request", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "only JSON format is supported at the moment", errResponse.Errors[0].Message)
}

func TestExportHandler_WithInvalidCollection_ReturnError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	filepath := t.TempDir() + "/test.json"
	cfg := client.BackupConfig{
		Filepath:    filepath,
		Format:      "json",
		Collections: []string{"invalid"},
	}
	b, err := json.Marshal(cfg)
	require.NoError(t, err)
	buf := bytes.NewBuffer(b)

	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           ExportPath,
		Body:           buf,
		ExpectedStatus: 400,
		ResponseData:   &errResponse,
	})
	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "collection does not exist: datastore: key not found")
	require.Equal(t, http.StatusBadRequest, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Bad Request", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "collection does not exist: datastore: key not found", errResponse.Errors[0].Message)
}

func TestExportHandler_WithBasicExportError_ReturnError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	db := mocks.NewDB(t)
	testError := errors.New("test error")
	db.EXPECT().BasicExport(mock.Anything, mock.Anything).Return(testError)

	filepath := t.TempDir() + "/test.json"
	cfg := client.BackupConfig{
		Filepath: filepath,
	}
	b, err := json.Marshal(cfg)
	require.NoError(t, err)
	buf := bytes.NewBuffer(b)

	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             db,
		Method:         "POST",
		Path:           ExportPath,
		Body:           buf,
		ExpectedStatus: 500,
		ResponseData:   &errResponse,
	})
	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "test error")
	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "test error", errResponse.Errors[0].Message)
}

// func TestExportHandler_WithGetCollectionError_ReturnError(t *testing.T) {
// 	t.Cleanup(CleanupEnv)
// 	env = "dev"
// 	db := mocks.NewDB(t)
// 	testError := errors.New("test error")
// 	db.EXPECT().GetAllCollections(mock.Anything).Return(nil, testError)

// 	errResponse := ErrorResponse{}
// 	testRequest(testOptions{
// 		Testing:        t,
// 		DB:             db,
// 		Method:         "GET",
// 		Path:           ExportPath,
// 		Body:           nil,
// 		ExpectedStatus: 500,
// 		ResponseData:   &errResponse,
// 	})
// 	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "test error")
// 	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
// 	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
// 	require.Equal(t, "test error", errResponse.Errors[0].Message)
// }

// func TestExportHandler_WithGetAllDockeysError_ReturnError(t *testing.T) {
// 	t.Cleanup(CleanupEnv)
// 	env = "dev"

// 	db := mocks.NewDB(t)
// 	testError := errors.New("test error")
// 	col := mocks.NewCollection(t)
// 	col.EXPECT().GetAllDocKeys(mock.Anything).Return(nil, testError)
// 	col.EXPECT().Schema().Return(client.SchemaDescription{Name: "test"})
// 	db.EXPECT().GetAllCollections(mock.Anything).Return([]client.Collection{col}, nil)

// 	errResponse := ErrorResponse{}
// 	testRequest(testOptions{
// 		Testing:        t,
// 		DB:             db,
// 		Method:         "GET",
// 		Path:           ExportPath,
// 		Body:           nil,
// 		ExpectedStatus: 500,
// 		ResponseData:   &errResponse,
// 	})
// 	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "test error")
// 	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
// 	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
// 	require.Equal(t, "test error", errResponse.Errors[0].Message)
// }

// func TestExportHandler_WithColGetError_ReturnError(t *testing.T) {
// 	t.Cleanup(CleanupEnv)
// 	env = "dev"

// 	db := mocks.NewDB(t)
// 	testError := errors.New("test error")
// 	col := mocks.NewCollection(t)
// 	keyCh := make(chan client.DocKeysResult, 2)
// 	keyCh <- client.DocKeysResult{}
// 	close(keyCh)
// 	col.EXPECT().GetAllDocKeys(mock.Anything).Return(keyCh, nil)
// 	col.EXPECT().Schema().Return(client.SchemaDescription{Name: "test"})
// 	col.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil, testError)
// 	db.EXPECT().GetAllCollections(mock.Anything).Return([]client.Collection{col}, nil)

// 	errResponse := ErrorResponse{}
// 	testRequest(testOptions{
// 		Testing:        t,
// 		DB:             db,
// 		Method:         "GET",
// 		Path:           ExportPath,
// 		Body:           nil,
// 		ExpectedStatus: 500,
// 		ResponseData:   &errResponse,
// 	})
// 	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "test error")
// 	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
// 	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
// 	require.Equal(t, "test error", errResponse.Errors[0].Message)
// }

func TestExportHandler_AllCollections_NoError(t *testing.T) {
	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	testLoadSchema(t, ctx, defra)

	col, err := defra.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"age": 31, "verified": true, "points": 90, "name": "Bob"}`))
	require.NoError(t, err)

	err = col.Create(ctx, doc)
	require.NoError(t, err)

	filepath := t.TempDir() + "/test.json"
	cfg := client.BackupConfig{
		Filepath: filepath,
	}
	b, err := json.Marshal(cfg)
	require.NoError(t, err)
	buf := bytes.NewBuffer(b)

	respBody := testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           ExportPath,
		Body:           buf,
		ExpectedStatus: 200,
	})

	b, err = os.ReadFile(filepath)
	require.NoError(t, err)

	require.Equal(
		t,
		`{"data":{"result":"success"}}`,
		string(respBody),
	)

	require.Equal(
		t,
		`{"User":[{"_key":"bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab","_newKey":"bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab","age":31,"name":"Bob","points":90,"verified":true}]}`,
		string(b),
	)
}

func TestExportHandler_UserCollection_NoError(t *testing.T) {
	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	testLoadSchema(t, ctx, defra)

	col, err := defra.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"age": 31, "verified": true, "points": 90, "name": "Bob"}`))
	require.NoError(t, err)

	err = col.Create(ctx, doc)
	require.NoError(t, err)

	filepath := t.TempDir() + "/test.json"
	cfg := client.BackupConfig{
		Filepath:    filepath,
		Collections: []string{"User"},
	}
	b, err := json.Marshal(cfg)
	require.NoError(t, err)
	buf := bytes.NewBuffer(b)

	respBody := testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           ExportPath,
		Body:           buf,
		ExpectedStatus: 200,
	})

	b, err = os.ReadFile(filepath)
	require.NoError(t, err)

	require.Equal(
		t,
		`{"data":{"result":"success"}}`,
		string(respBody),
	)

	require.Equal(
		t,
		`{"User":[{"_key":"bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab","_newKey":"bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab","age":31,"name":"Bob","points":90,"verified":true}]}`,
		string(b),
	)
}

func TestExportHandler_UserCollectionWithModifiedDoc_NoError(t *testing.T) {
	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	testLoadSchema(t, ctx, defra)

	col, err := defra.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"age": 31, "verified": true, "points": 90, "name": "Bob"}`))
	require.NoError(t, err)

	err = col.Create(ctx, doc)
	require.NoError(t, err)

	err = doc.Set("points", 1000)
	require.NoError(t, err)

	err = col.Update(ctx, doc)
	require.NoError(t, err)

	filepath := t.TempDir() + "/test.json"
	cfg := client.BackupConfig{
		Filepath:    filepath,
		Collections: []string{"User"},
	}
	b, err := json.Marshal(cfg)
	require.NoError(t, err)
	buf := bytes.NewBuffer(b)

	respBody := testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           ExportPath,
		Body:           buf,
		ExpectedStatus: 200,
	})

	b, err = os.ReadFile(filepath)
	require.NoError(t, err)

	require.Equal(
		t,
		`{"data":{"result":"success"}}`,
		string(respBody),
	)

	require.Equal(
		t,
		`{"User":[{"_key":"bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab","_newKey":"bae-36697142-d46a-57b1-b25e-6336706854ea","age":31,"name":"Bob","points":1000,"verified":true}]}`,
		string(b),
	)
}

func TestImportHandler_WithNoDB_NoDatabaseAvailableError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             nil,
		Method:         "POST",
		Path:           ImportPath,
		Body:           nil,
		ExpectedStatus: 500,
		ResponseData:   &errResponse,
	})
	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "no database available")
	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "no database available", errResponse.Errors[0].Message)
}

func TestImportHandler_WithWrongPayloadFormat_UnmarshalError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"

	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	buf := bytes.NewBuffer([]byte(`[]`))

	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           ImportPath,
		Body:           buf,
		ExpectedStatus: 400,
		ResponseData:   &errResponse,
	})
	require.Contains(
		t,
		errResponse.Errors[0].Extensions.Stack,
		"json: cannot unmarshal array into Go value of type client.BackupConfig",
	)
	require.Equal(t, http.StatusBadRequest, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Bad Request", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(
		t,
		"unmarshal error: json: cannot unmarshal array into Go value of type client.BackupConfig",
		errResponse.Errors[0].Message,
	)
}

func TestImportHandler_WithInvalidFilepath_ReturnError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"

	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	filepath := t.TempDir() + "/some/test.json"
	cfg := client.BackupConfig{
		Filepath: filepath,
	}
	b, err := json.Marshal(cfg)
	require.NoError(t, err)
	buf := bytes.NewBuffer(b)

	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           ImportPath,
		Body:           buf,
		ExpectedStatus: 400,
		ResponseData:   &errResponse,
	})
	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "invalid file path")
	require.Equal(t, http.StatusBadRequest, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Bad Request", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "invalid file path", errResponse.Errors[0].Message)
}

func TestImportHandler_WithDBClosed_DatastoreClosedError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"

	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defra.Close(ctx)

	filepath := t.TempDir() + "/test.json"
	cfg := client.BackupConfig{
		Filepath: filepath,
	}
	b, err := json.Marshal(cfg)
	require.NoError(t, err)
	buf := bytes.NewBuffer(b)

	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           ImportPath,
		Body:           buf,
		ExpectedStatus: 500,
		ResponseData:   &errResponse,
	})
	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "datastore closed")
	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "datastore closed", errResponse.Errors[0].Message)
}

func TestImportHandler_WithUnknownCollection_KeyNotFoundError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"

	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	filepath := t.TempDir() + "/test.json"
	err := os.WriteFile(
		filepath,
		[]byte(`{"User":[{"_key":"bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab","age":31,"name":"Bob","points":90,"verified":true}]}`),
		0644,
	)
	require.NoError(t, err)

	cfg := client.BackupConfig{
		Filepath: filepath,
	}

	b, err := json.Marshal(cfg)
	require.NoError(t, err)
	buf := bytes.NewBuffer(b)

	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           ImportPath,
		Body:           buf,
		ExpectedStatus: 500,
		ResponseData:   &errResponse,
	})
	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "datastore: key not found")
	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "datastore: key not found", errResponse.Errors[0].Message)
}

// type mockStore struct {
// 	client.DB
// 	datastore.Txn
// }

// func TestImportHandler_WithTxnCommitError_ReturnError(t *testing.T) {
// 	t.Cleanup(CleanupEnv)
// 	env = "dev"

// 	testError := errors.New("test error")
// 	db := mocks.NewDB(t)
// 	txn := dsMocks.NewTxn(t)
// 	store := mockStore{
// 		DB:  db,
// 		Txn: txn,
// 	}
// 	db.EXPECT().NewTxn(mock.Anything, mock.Anything).Return(txn, nil)
// 	db.EXPECT().WithTxn(mock.Anything).Return(store)
// 	txn.EXPECT().Discard(mock.Anything).Return()
// 	txn.EXPECT().Commit(mock.Anything).Return(testError)

// 	buf := bytes.NewBuffer([]byte(`{}`))

// 	errResponse := ErrorResponse{}
// 	testRequest(testOptions{
// 		Testing:        t,
// 		DB:             db,
// 		Method:         "POST",
// 		Path:           ImportPath,
// 		Body:           buf,
// 		ExpectedStatus: 500,
// 		ResponseData:   &errResponse,
// 	})
// 	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "test error")
// 	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
// 	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
// 	require.Equal(t, "test error", errResponse.Errors[0].Message)
// }

func TestImportHandler_UserCollection_NoError(t *testing.T) {
	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	testLoadSchema(t, ctx, defra)

	filepath := t.TempDir() + "/test.json"
	err := os.WriteFile(
		filepath,
		[]byte(`{"User":[{"_key":"bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab","_newKey":"bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab","age":31,"name":"Bob","points":90,"verified":true}]}`),
		0644,
	)
	require.NoError(t, err)

	cfg := client.BackupConfig{
		Filepath: filepath,
	}

	b, err := json.Marshal(cfg)
	require.NoError(t, err)
	buf := bytes.NewBuffer(b)

	resp := DataResponse{}
	_ = testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           ImportPath,
		Body:           buf,
		ExpectedStatus: 200,
		ResponseData:   &resp,
	})

	switch v := resp.Data.(type) {
	case map[string]any:
		require.Equal(t, "success", v["result"])
	default:
		t.Fatalf("data should be of type map[string]any but got %T", resp.Data)
	}

	doc, err := client.NewDocFromJSON([]byte(`{"age": 31, "verified": true, "points": 90, "name": "Bob"}`))
	require.NoError(t, err)

	col, err := defra.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	importedDoc, err := col.Get(ctx, doc.Key(), false)
	require.NoError(t, err)

	require.Equal(t, doc.Key().String(), importedDoc.Key().String())
}

func TestImportHandler_WithExistingDoc_DocumentExistError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"

	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	testLoadSchema(t, ctx, defra)

	col, err := defra.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"age": 31, "verified": true, "points": 90, "name": "Bob"}`))
	require.NoError(t, err)

	err = col.Create(ctx, doc)
	require.NoError(t, err)

	filepath := t.TempDir() + "/test.json"
	err = os.WriteFile(
		filepath,
		[]byte(`{"User":[{"_key":"bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab","_newKey":"bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab","age":31,"name":"Bob","points":90,"verified":true}]}`),
		0644,
	)
	require.NoError(t, err)

	cfg := client.BackupConfig{
		Filepath: filepath,
	}

	b, err := json.Marshal(cfg)
	require.NoError(t, err)
	buf := bytes.NewBuffer(b)

	errResponse := ErrorResponse{}
	_ = testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           ImportPath,
		QueryParams:    map[string]string{"collections": "User"},
		Body:           buf,
		ExpectedStatus: 500,
		ResponseData:   &errResponse,
	})

	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "a document with the given dockey already exists")
	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(
		t,
		"a document with the given dockey already exists. DocKey: bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab",
		errResponse.Errors[0].Message,
	)
}

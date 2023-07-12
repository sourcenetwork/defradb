package http

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/mocks"
	"github.com/sourcenetwork/defradb/datastore"
	dsMocks "github.com/sourcenetwork/defradb/datastore/mocks"
	"github.com/sourcenetwork/defradb/errors"
)

func TestExportHandler_WithNoDB_NoDatabaseAvailableError1(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	db := mocks.NewDB(t)
	testError := errors.New("test error")
	db.EXPECT().GetAllCollections(mock.Anything).Return(nil, testError)

	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             db,
		Method:         "GET",
		Path:           ExportPath,
		Body:           nil,
		ExpectedStatus: 500,
		ResponseData:   &errResponse,
	})
	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "test error")
	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "test error", errResponse.Errors[0].Message)
}

func TestExportHandler_WithNoDB_NoDatabaseAvailableError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             nil,
		Method:         "GET",
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

func TestExportHandler_WithGetCollectionError_ReturnError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	db := mocks.NewDB(t)
	testError := errors.New("test error")
	db.EXPECT().GetAllCollections(mock.Anything).Return(nil, testError)

	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             db,
		Method:         "GET",
		Path:           ExportPath,
		Body:           nil,
		ExpectedStatus: 500,
		ResponseData:   &errResponse,
	})
	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "test error")
	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "test error", errResponse.Errors[0].Message)
}

func TestExportHandler_WithGetAllDockeysError_ReturnError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"

	db := mocks.NewDB(t)
	testError := errors.New("test error")
	col := mocks.NewCollection(t)
	col.EXPECT().GetAllDocKeys(mock.Anything).Return(nil, testError)
	col.EXPECT().Schema().Return(client.SchemaDescription{Name: "test"})
	db.EXPECT().GetAllCollections(mock.Anything).Return([]client.Collection{col}, nil)

	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             db,
		Method:         "GET",
		Path:           ExportPath,
		Body:           nil,
		ExpectedStatus: 500,
		ResponseData:   &errResponse,
	})
	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "test error")
	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "test error", errResponse.Errors[0].Message)
}

func TestExportHandler_WithColGetError_ReturnError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"

	db := mocks.NewDB(t)
	testError := errors.New("test error")
	col := mocks.NewCollection(t)
	keyCh := make(chan client.DocKeysResult, 2)
	keyCh <- client.DocKeysResult{}
	close(keyCh)
	col.EXPECT().GetAllDocKeys(mock.Anything).Return(keyCh, nil)
	col.EXPECT().Schema().Return(client.SchemaDescription{Name: "test"})
	col.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil, testError)
	db.EXPECT().GetAllCollections(mock.Anything).Return([]client.Collection{col}, nil)

	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             db,
		Method:         "GET",
		Path:           ExportPath,
		Body:           nil,
		ExpectedStatus: 500,
		ResponseData:   &errResponse,
	})
	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "test error")
	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "test error", errResponse.Errors[0].Message)
}

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

	respBody := testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "GET",
		Path:           ExportPath,
		Body:           nil,
		ExpectedStatus: 200,
	})

	require.Equal(
		t,
		`{"data":{"User":[{"_key":"bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab","_newKey":"bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab","age":31,"name":"Bob","points":90,"verified":true}]}}`,
		string(respBody),
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

	respBody := testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "GET",
		Path:           ExportPath,
		QueryParams:    map[string]string{"collections": "User"},
		Body:           nil,
		ExpectedStatus: 200,
	})

	require.Equal(
		t,
		`{"data":{"User":[{"_key":"bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab","_newKey":"bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab","age":31,"name":"Bob","points":90,"verified":true}]}}`,
		string(respBody),
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

	respBody := testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "GET",
		Path:           ExportPath,
		QueryParams:    map[string]string{"collections": "User"},
		Body:           nil,
		ExpectedStatus: 200,
	})

	require.Equal(
		t,
		`{"data":{"User":[{"_key":"bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab","_newKey":"bae-36697142-d46a-57b1-b25e-6336706854ea","age":31,"name":"Bob","points":1000,"verified":true}]}}`,
		string(respBody),
	)
}

func TestExportHandler_InvalidCollection_KeyNotFoundError(t *testing.T) {
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

	errResponse := ErrorResponse{}
	_ = testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "GET",
		Path:           ExportPath,
		QueryParams:    map[string]string{"collections": "Invalid"},
		Body:           nil,
		ExpectedStatus: 400,
		ResponseData:   &errResponse,
	})

	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "datastore: key not found")
	require.Equal(t, http.StatusBadRequest, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Bad Request", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "datastore: key not found", errResponse.Errors[0].Message)
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

	buf := bytes.NewBuffer([]byte(`[{"_key":"bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab","age":31,"name":"Bob","points":90,"verified":true}]`))

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
		"json: cannot unmarshal array into Go value of type map[string][]map[string]interface {}",
	)
	require.Equal(t, http.StatusBadRequest, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Bad Request", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(
		t,
		"unmarshal error: json: cannot unmarshal array into Go value of type map[string][]map[string]interface {}",
		errResponse.Errors[0].Message,
	)
}

func TestImportHandler_WithDBClosed_DatastoreClosedError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"

	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defra.Close(ctx)

	buf := bytes.NewBuffer([]byte(`{"User":[{"_key":"bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab","age":31,"name":"Bob","points":90,"verified":true}]}`))

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

	buf := bytes.NewBuffer([]byte(`{"User":[{"_key":"bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab","age":31,"name":"Bob","points":90,"verified":true}]}`))

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
	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "datastore: key not found")
	require.Equal(t, http.StatusBadRequest, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Bad Request", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "datastore: key not found", errResponse.Errors[0].Message)
}

func TestImportHandler_WithInvalidDockeyFormat_KeyNotFoundError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"

	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	testLoadSchema(t, ctx, defra)

	buf := bytes.NewBuffer([]byte(`{"User":[{"_key":"91171025-ed21-50e3-b0dc-e31bccdfa1ab","_newKey":"91171025-ed21-50e3-b0dc-e31bccdfa1ab","age":31,"name":"Bob","points":90,"verified":true}]}`))

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
	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "selected encoding not supported")
	require.Equal(t, http.StatusBadRequest, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Bad Request", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "selected encoding not supported", errResponse.Errors[0].Message)
}

type mockStore struct {
	client.DB
	datastore.Txn
}

func TestImportHandler_WithTxnCommitError_ReturnError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"

	testError := errors.New("test error")
	db := mocks.NewDB(t)
	txn := dsMocks.NewTxn(t)
	store := mockStore{
		DB:  db,
		Txn: txn,
	}
	db.EXPECT().NewTxn(mock.Anything, mock.Anything).Return(txn, nil)
	db.EXPECT().WithTxn(mock.Anything).Return(store)
	txn.EXPECT().Discard(mock.Anything).Return()
	txn.EXPECT().Commit(mock.Anything).Return(testError)

	buf := bytes.NewBuffer([]byte(`{}`))

	errResponse := ErrorResponse{}
	testRequest(testOptions{
		Testing:        t,
		DB:             db,
		Method:         "POST",
		Path:           ImportPath,
		Body:           buf,
		ExpectedStatus: 500,
		ResponseData:   &errResponse,
	})
	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "test error")
	require.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Internal Server Error", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(t, "test error", errResponse.Errors[0].Message)
}

func TestImportHandler_UserCollection_NoError(t *testing.T) {
	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	testLoadSchema(t, ctx, defra)

	col, err := defra.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	buf := bytes.NewBuffer([]byte(`{"User":[{"_key":"bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab","_newKey":"bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab","age":31,"name":"Bob","points":90,"verified":true}]}`))

	resp := DataResponse{}
	_ = testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           ImportPath,
		QueryParams:    map[string]string{"collections": "User"},
		Body:           buf,
		ExpectedStatus: 200,
		ResponseData:   &resp,
	})

	switch v := resp.Data.(type) {
	case map[string]any:
		require.Equal(t, "ok", v["response"])
	default:
		t.Fatalf("data should be of type map[string]any but got %T", resp.Data)
	}

	doc, err := client.NewDocFromJSON([]byte(`{"age": 31, "verified": true, "points": 90, "name": "Bob"}`))
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

	buf := bytes.NewBuffer([]byte(`{"User":[{"_key":"bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab","_newKey":"bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab","age":31,"name":"Bob","points":90,"verified":true}]}`))

	errResponse := ErrorResponse{}
	_ = testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           ImportPath,
		QueryParams:    map[string]string{"collections": "User"},
		Body:           buf,
		ExpectedStatus: 400,
		ResponseData:   &errResponse,
	})

	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "a document with the given dockey already exists")
	require.Equal(t, http.StatusBadRequest, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Bad Request", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(
		t,
		"a document with the given dockey already exists. DocKey: bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab",
		errResponse.Errors[0].Message,
	)
}

func TestImportHandler_WithMisingNewKey_MissingNewKeyError(t *testing.T) {
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

	buf := bytes.NewBuffer([]byte(`{"User":[{"_key":"bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab","age":31,"name":"Bob","points":90,"verified":true}]}`))

	errResponse := ErrorResponse{}
	_ = testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           ImportPath,
		QueryParams:    map[string]string{"collections": "User"},
		Body:           buf,
		ExpectedStatus: 400,
		ResponseData:   &errResponse,
	})

	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "missing _newKey for imported doc")
	require.Equal(t, http.StatusBadRequest, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Bad Request", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(
		t,
		"missing _newKey for imported doc",
		errResponse.Errors[0].Message,
	)
}

func TestImportHandler_WithCBORFormat_NoError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"

	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	testLoadSchema(t, ctx, defra)

	b, err := os.ReadFile("./handlerfuncs_export_test.cbor")
	require.NoError(t, err)
	buf := bytes.NewBuffer(b)

	resp := DataResponse{}
	_ = testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           ImportPath,
		QueryParams:    map[string]string{"collections": "User"},
		Headers:        map[string]string{"Content-Type": "application/octet-stream"},
		Body:           buf,
		ExpectedStatus: 200,
		ResponseData:   &resp,
	})

	switch v := resp.Data.(type) {
	case map[string]any:
		require.Equal(t, "ok", v["response"])
	default:
		t.Fatalf("data should be of type map[string]any but got %T", resp.Data)
	}

	col, err := defra.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"age": 41, "name": "John"}`))
	require.NoError(t, err)

	importedDoc, err := col.Get(ctx, doc.Key(), false)
	require.NoError(t, err)

	require.Equal(t, doc.Key().String(), importedDoc.Key().String())
}

func TestImportHandler_WithCBORFormatAndInvalidContent_EndOfFileError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"

	ctx := context.Background()
	defra := testNewInMemoryDB(t, ctx)
	defer defra.Close(ctx)

	testLoadSchema(t, ctx, defra)

	buf := bytes.NewBuffer([]byte(`{"User":[{"_key":"bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab","age":31,"name":"Bob","points":90,"verified":true}]}`))

	errResponse := ErrorResponse{}
	_ = testRequest(testOptions{
		Testing:        t,
		DB:             defra,
		Method:         "POST",
		Path:           ImportPath,
		QueryParams:    map[string]string{"collections": "User"},
		Headers:        map[string]string{"Content-Type": "application/octet-stream"},
		Body:           buf,
		ExpectedStatus: 400,
		ResponseData:   &errResponse,
	})

	require.Contains(t, errResponse.Errors[0].Extensions.Stack, "EOF")
	require.Equal(t, http.StatusBadRequest, errResponse.Errors[0].Extensions.Status)
	require.Equal(t, "Bad Request", errResponse.Errors[0].Extensions.HTTPError)
	require.Equal(
		t,
		"EOF",
		errResponse.Errors[0].Message,
	)
}

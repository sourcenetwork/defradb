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
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/mocks"
	"github.com/sourcenetwork/defradb/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func addDBToContext(t *testing.T, req *http.Request, db *mocks.DB) *http.Request {
	if db == nil {
		db = mocks.NewDB(t)
	}
	ctx := context.WithValue(req.Context(), ctxDB{}, db)
	return req.WithContext(ctx)
}

func TestCreateIndexHandler_IfNoDBInContext_ReturnError(t *testing.T) {
	handler := http.HandlerFunc(createIndexHandler)
	assert.HTTPBodyContains(t, handler, "POST", IndexPath, nil, "no database available")
}

func TestCreateIndexHandler_IfFailsToParseParams_ReturnError(t *testing.T) {
	req, err := http.NewRequest("POST", IndexPath, bytes.NewBuffer([]byte("invalid map")))
	if err != nil {
		t.Fatal(err)
	}
	req = addDBToContext(t, req, nil)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(createIndexHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code, "handler returned wrong status code")
	assert.Contains(t, rr.Body.String(), "invalid character", "handler returned unexpected body")
}

func TestCreateIndexHandler_IfFailsToGetCollection_ReturnError(t *testing.T) {
	testError := errors.New("test error")
	db := mocks.NewDB(t)
	db.EXPECT().GetCollectionByName(mock.Anything, mock.Anything).Return(nil, testError)

	req, err := http.NewRequest("POST", IndexPath, bytes.NewBuffer([]byte(`{}`)))
	if err != nil {
		t.Fatal(err)
	}

	req = addDBToContext(t, req, db)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(createIndexHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "handler returned wrong status code")
	assert.Contains(t, rr.Body.String(), testError.Error())
}

func TestCreateIndexHandler_IfFailsToCreateIndex_ReturnError(t *testing.T) {
	testError := errors.New("test error")
	col := mocks.NewCollection(t)
	col.EXPECT().CreateIndex(mock.Anything, mock.Anything).
		Return(client.IndexDescription{}, testError)

	db := mocks.NewDB(t)
	db.EXPECT().GetCollectionByName(mock.Anything, mock.Anything).Return(col, nil)

	req, err := http.NewRequest("POST", IndexPath, bytes.NewBuffer([]byte(`{}`)))
	if err != nil {
		t.Fatal(err)
	}
	req = addDBToContext(t, req, db)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(createIndexHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "handler returned wrong status code")
	assert.Contains(t, rr.Body.String(), testError.Error())
}

func TestDropIndexHandler_IfNoDBInContext_ReturnError(t *testing.T) {
	handler := http.HandlerFunc(dropIndexHandler)
	assert.HTTPBodyContains(t, handler, "DELETE", IndexPath, nil, "no database available")
}

func TestDropIndexHandler_IfFailsToParseParams_ReturnError(t *testing.T) {
	req, err := http.NewRequest("DELETE", IndexPath, bytes.NewBuffer([]byte("invalid map")))
	if err != nil {
		t.Fatal(err)
	}
	req = addDBToContext(t, req, nil)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(dropIndexHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code, "handler returned wrong status code")
	assert.Contains(t, rr.Body.String(), "invalid character", "handler returned unexpected body")
}

func TestDropIndexHandler_IfFailsToGetCollection_ReturnError(t *testing.T) {
	testError := errors.New("test error")
	db := mocks.NewDB(t)
	db.EXPECT().GetCollectionByName(mock.Anything, mock.Anything).Return(nil, testError)

	req, err := http.NewRequest("DELETE", IndexPath, bytes.NewBuffer([]byte(`{}`)))
	if err != nil {
		t.Fatal(err)
	}

	req = addDBToContext(t, req, db)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(dropIndexHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "handler returned wrong status code")
	assert.Contains(t, rr.Body.String(), testError.Error())
}

func TestDropIndexHandler_IfFailsToDropIndex_ReturnError(t *testing.T) {
	testError := errors.New("test error")
	col := mocks.NewCollection(t)
	col.EXPECT().DropIndex(mock.Anything, mock.Anything).Return(testError)

	db := mocks.NewDB(t)
	db.EXPECT().GetCollectionByName(mock.Anything, mock.Anything).Return(col, nil)

	req, err := http.NewRequest("DELETE", IndexPath, bytes.NewBuffer([]byte(`{}`)))
	if err != nil {
		t.Fatal(err)
	}
	req = addDBToContext(t, req, db)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(dropIndexHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "handler returned wrong status code")
	assert.Contains(t, rr.Body.String(), testError.Error())
}

func TestListIndexHandler_IfNoDBInContext_ReturnError(t *testing.T) {
	handler := http.HandlerFunc(listIndexHandler)
	assert.HTTPBodyContains(t, handler, "GET", IndexPath, nil, "no database available")
}

func TestListIndexHandler_IfFailsToGetAllIndexes_ReturnError(t *testing.T) {
	testError := errors.New("test error")
	db := mocks.NewDB(t)
	db.EXPECT().GetAllIndexes(mock.Anything).Return(nil, testError)

	req, err := http.NewRequest("GET", IndexPath, bytes.NewBuffer([]byte(`{}`)))
	if err != nil {
		t.Fatal(err)
	}

	req = addDBToContext(t, req, db)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(listIndexHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "handler returned wrong status code")
	assert.Contains(t, rr.Body.String(), testError.Error())
}

func TestListIndexHandler_IfFailsToGetCollection_ReturnError(t *testing.T) {
	testError := errors.New("test error")
	db := mocks.NewDB(t)
	db.EXPECT().GetCollectionByName(mock.Anything, mock.Anything).Return(nil, testError)

	u, _ := url.Parse("http://defradb.com" + IndexPath)
	params := url.Values{}
	params.Add("collection", "testCollection")
	u.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		t.Fatal(err)
	}

	req = addDBToContext(t, req, db)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(listIndexHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "handler returned wrong status code")
	assert.Contains(t, rr.Body.String(), testError.Error())
}

func TestListIndexHandler_IfFailsToCollectionGetIndexes_ReturnError(t *testing.T) {
	testError := errors.New("test error")
	col := mocks.NewCollection(t)
	col.EXPECT().GetIndexes(mock.Anything).Return(nil, testError)

	db := mocks.NewDB(t)
	db.EXPECT().GetCollectionByName(mock.Anything, mock.Anything).Return(col, nil)

	u, _ := url.Parse("http://defradb.com" + IndexPath)
	params := url.Values{}
	params.Add("collection", "testCollection")
	u.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		t.Fatal(err)
	}
	req = addDBToContext(t, req, db)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(listIndexHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "handler returned wrong status code")
	assert.Contains(t, rr.Body.String(), testError.Error())
}

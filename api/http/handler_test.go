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
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/pkg/errors"
	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v3"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/stretchr/testify/assert"
)

func TestNewHandlerWithLogger(t *testing.T) {
	h := newHandler(nil)

	dir := t.TempDir()

	// send logs to temp file so we can inspect it
	logFile := path.Join(dir, "http_test.log")
	log.ApplyConfig(logging.Config{
		EncoderFormat: logging.NewEncoderFormatOption(logging.JSON),
		OutputPaths:   []string{logFile},
	})

	req, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()

	loggerMiddleware(h.handle(pingHandler)).ServeHTTP(rec, req)
	assert.Equal(t, 200, rec.Result().StatusCode)

	// inspect the log file
	kv, err := readLog(logFile)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "defra.http", kv["logger"])

}

func TestGetJSON(t *testing.T) {
	var obj struct {
		Name string
	}

	jsonStr := []byte(`
		{
			"Name": "John Doe"
		}
	`)

	req, err := http.NewRequest("POST", "/ping", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}

	err = getJSON(req, &obj)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "John Doe", obj.Name)

}

func TestGetJSONWithError(t *testing.T) {
	var obj struct {
		Name string
	}

	jsonStr := []byte(`
		{
			"Name": 10
		}
	`)

	req, err := http.NewRequest("POST", "/ping", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}

	err = getJSON(req, &obj)
	assert.Error(t, err)
}

func TestSendJSONWithNoErrors(t *testing.T) {
	obj := struct {
		Name string
	}{
		Name: "John Doe",
	}

	rec := httptest.NewRecorder()

	sendJSON(context.Background(), rec, obj, 200)

	body, err := ioutil.ReadAll(rec.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []byte("{\"Name\":\"John Doe\"}"), body)
}

func TestSendJSONWithMarshallFailure(t *testing.T) {
	rec := httptest.NewRecorder()

	sendJSON(context.Background(), rec, math.Inf(1), 200)

	assert.Equal(t, http.StatusInternalServerError, rec.Result().StatusCode)
}

type loggerTest struct {
	loggingResponseWriter
}

func (lt *loggerTest) Write(b []byte) (int, error) {
	return 0, errors.New("this write will fail")
}

func TestSendJSONWithMarshallFailureAndWriteFailer(t *testing.T) {
	rec := httptest.NewRecorder()
	lrw := loggerTest{}
	lrw.ResponseWriter = rec

	sendJSON(context.Background(), &lrw, math.Inf(1), 200)

	assert.Equal(t, http.StatusInternalServerError, rec.Result().StatusCode)
}

func TestSendJSONWithWriteFailure(t *testing.T) {
	obj := struct {
		Name string
	}{
		Name: "John Doe",
	}

	rec := httptest.NewRecorder()
	lrw := loggerTest{}
	lrw.ResponseWriter = rec

	sendJSON(context.Background(), &lrw, obj, 200)

	assert.Equal(t, http.StatusInternalServerError, lrw.statusCode)
}

func TestDbFromContext(t *testing.T) {
	_, err := dbFromContext(context.Background())
	assert.Error(t, err)

	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	if err != nil {
		t.Fatal(err)
	}

	var options []db.Option
	ctx := context.Background()

	defra, err := db.NewDB(ctx, rootstore, options...)
	if err != nil {
		t.Fatal(err)
	}

	reqCtx := context.WithValue(ctx, ctxDB{}, defra)

	_, err = dbFromContext(reqCtx)
	assert.NoError(t, err)
}

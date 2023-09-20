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
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v4"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/logging"
)

func TestSimpleDataResponse(t *testing.T) {
	resp := simpleDataResponse("key", "value", "key2", "value2")
	switch v := resp.Data.(type) {
	case map[string]any:
		assert.Equal(t, "value", v["key"])
		assert.Equal(t, "value2", v["key2"])
	default:
		t.Fatalf("data should be of type map[string]any but got %T", resp.Data)
	}

	resp2 := simpleDataResponse("key", "value", "key2")
	switch v := resp2.Data.(type) {
	case map[string]any:
		assert.Equal(t, "value", v["key"])
		assert.Equal(t, nil, v["key2"])
	default:
		t.Fatalf("data should be of type map[string]any but got %T", resp.Data)
	}

	resp3 := simpleDataResponse("key", "value", 2, "value2")
	switch v := resp3.Data.(type) {
	case map[string]any:
		assert.Equal(t, "value", v["key"])
		assert.Equal(t, nil, v["2"])
	default:
		t.Fatalf("data should be of type map[string]any but got %T", resp.Data)
	}
}

func TestNewHandlerWithLogger(t *testing.T) {
	h := newHandler(nil, serverOptions{})

	dir := t.TempDir()

	// send logs to temp file so we can inspect it
	logFile := path.Join(dir, "http_test.log")
	log.ApplyConfig(logging.Config{
		EncoderFormat: logging.NewEncoderFormatOption(logging.JSON),
		OutputPaths:   []string{logFile},
	})

	req, err := http.NewRequest("GET", PingPath, nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	lrw := newLoggingResponseWriter(rec)
	h.ServeHTTP(lrw, req)
	assert.Equal(t, 200, rec.Result().StatusCode)

	// inspect the log file
	kv, err := readLog(logFile)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "http", kv["logger"])
}

func TestGetJSON(t *testing.T) {
	var obj struct {
		Name string
	}

	jsonStr := `
{
	"Name": "John Doe"
}`

	req, err := http.NewRequest("POST", "/ping", bytes.NewBuffer([]byte(jsonStr)))
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

	jsonStr := `
{
	"Name": 10
}`

	req, err := http.NewRequest("POST", "/ping", bytes.NewBuffer([]byte(jsonStr)))
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

	body, err := io.ReadAll(rec.Result().Body)
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

func TestCORSRequest(t *testing.T) {
	cases := []struct {
		name       string
		method     string
		reqHeaders map[string]string
		resHeaders map[string]string
	}{
		{
			"DisallowedOrigin",
			"OPTIONS",
			map[string]string{
				"Origin": "https://notsource.network",
			},
			map[string]string{
				"Vary": "Origin",
			},
		},
		{
			"AllowedOrigin",
			"OPTIONS",
			map[string]string{
				"Origin": "https://source.network",
			},
			map[string]string{
				"Access-Control-Allow-Origin": "https://source.network",
				"Vary":                        "Origin",
			},
		},
	}

	s := NewServer(nil, WithAllowedOrigins("https://source.network"))

	for i := range cases {
		c := cases[i]
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest(c.method, PingPath, nil)
			if err != nil {
				t.Fatal(err)
			}

			for header, value := range c.reqHeaders {
				req.Header.Add(header, value)
			}

			rec := httptest.NewRecorder()

			s.Handler.ServeHTTP(rec, req)

			for header, value := range c.resHeaders {
				assert.Equal(t, value, rec.Result().Header.Get(header))
			}
		})
	}
}

func TestTLSRequestResponseHeader(t *testing.T) {
	cases := []struct {
		name       string
		method     string
		reqHeaders map[string]string
		resHeaders map[string]string
	}{
		{
			"TLSHeader",
			"GET",
			map[string]string{},
			map[string]string{
				"Strict-Transport-Security": "max-age=63072000; includeSubDomains",
			},
		},
	}
	dir := t.TempDir()

	s := NewServer(nil, WithTLS(), WithAddress("example.com"), WithRootDir(dir))

	for i := range cases {
		c := cases[i]
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest(c.method, PingPath, nil)
			if err != nil {
				t.Fatal(err)
			}

			for header, value := range c.reqHeaders {
				req.Header.Add(header, value)
			}

			rec := httptest.NewRecorder()

			s.Handler.ServeHTTP(rec, req)

			for header, value := range c.resHeaders {
				assert.Equal(t, value, rec.Result().Header.Get(header))
			}
		})
	}
}

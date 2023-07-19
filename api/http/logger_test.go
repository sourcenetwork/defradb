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
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/logging"
)

func TestNewLoggingResponseWriterLogger(t *testing.T) {
	rec := httptest.NewRecorder()
	lrw := newLoggingResponseWriter(rec)

	lrw.WriteHeader(400)
	assert.Equal(t, 400, lrw.statusCode)

	content := "Hello world!"

	length, err := lrw.Write([]byte(content))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, length, lrw.contentLength)
	assert.Equal(t, rec.Body.String(), content)
}

func TestLogginResponseWriterWriteWithChunks(t *testing.T) {
	rec := httptest.NewRecorder()
	lrw := newLoggingResponseWriter(rec)

	content := "Hello world!"
	contentLength := len(content)

	lrw.Header().Set("Content-Length", strconv.Itoa(contentLength))

	length1, err := lrw.Write([]byte(content[:contentLength/2]))
	if err != nil {
		t.Fatal(err)
	}

	length2, err := lrw.Write([]byte(content[contentLength/2:]))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, contentLength, length1+length2)
	assert.Equal(t, rec.Body.String(), content)
}

func TestLoggerKeyValueOutput(t *testing.T) {
	dir := t.TempDir()

	// send logs to temp file so we can inspect it
	logFile := path.Join(dir, "http_test.log")

	req, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec2 := httptest.NewRecorder()

	log.ApplyConfig(logging.Config{
		EncoderFormat: logging.NewEncoderFormatOption(logging.JSON),
		OutputPaths:   []string{logFile},
	})
	loggerMiddleware(http.HandlerFunc(pingHandler)).ServeHTTP(rec2, req)
	assert.Equal(t, 200, rec2.Result().StatusCode)

	// inspect the log file
	kv, err := readLog(logFile)
	if err != nil {
		t.Fatal(err)
	}

	// check that everything is as expected
	assert.Equal(t, "{\"data\":{\"response\":\"pong\"}}", rec2.Body.String())
	assert.Equal(t, "INFO", kv["level"])
	assert.Equal(t, "http", kv["logger"])
	assert.Equal(t, "Request", kv["msg"])
	assert.Equal(t, "GET", kv["Method"])
	assert.Equal(t, "/ping", kv["Path"])
	assert.Equal(t, float64(200), kv["Status"])
	assert.Equal(t, float64(28), kv["LengthBytes"])
}

func readLog(path string) (map[string]any, error) {
	// inspect the log file
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	scanner := bufio.NewScanner(f)
	scanner.Scan()
	logLine := scanner.Text()

	kv := make(map[string]any)
	err = json.Unmarshal([]byte(logLine), &kv)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return kv, nil
}

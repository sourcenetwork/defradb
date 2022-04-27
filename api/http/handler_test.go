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
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/sourcenetwork/defradb/logging"
	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	h := NewHandler(nil, nil)

	dir := t.TempDir()

	// send logs to temp file so we can inspect it
	logFile := path.Join(dir, "http_test.log")
	h.ApplyConfig(logging.Config{
		EncoderFormat: logging.NewEncoderFormatOption(logging.JSON),
		OutputPaths:   []string{logFile},
	})

	req, err := http.NewRequest("GET", "/ping", nil)
	assert.NoError(t, err)

	rec2 := httptest.NewRecorder()

	h.loggerMiddleware(h.handle(ping)).ServeHTTP(rec2, req)
	assert.Equal(t, 200, rec2.Result().StatusCode)

	// inspect the log file
	kv, err := readLog(logFile)
	assert.NoError(t, err)

	assert.Equal(t, "defra.http", kv["logger"])

}

func TestHandlerWithConfig(t *testing.T) {
	h := NewHandler(nil, &HandlerConfig{
		Logger: withLogger(logging.MustNewLogger("defra.http.test")),
	})

	dir := t.TempDir()

	// send logs to temp file so we can inspect it
	logFile := path.Join(dir, "http_test.log")
	h.ApplyConfig(logging.Config{
		EncoderFormat: logging.NewEncoderFormatOption(logging.JSON),
		OutputPaths:   []string{logFile},
	})

	req, err := http.NewRequest("GET", "/ping", nil)
	assert.NoError(t, err)

	rec2 := httptest.NewRecorder()

	h.loggerMiddleware(h.handle(ping)).ServeHTTP(rec2, req)
	assert.Equal(t, 200, rec2.Result().StatusCode)

	// inspect the log file
	kv, err := readLog(logFile)
	assert.NoError(t, err)

	assert.Equal(t, "defra.http.test", kv["logger"])

}

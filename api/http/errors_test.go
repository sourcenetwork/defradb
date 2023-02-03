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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func CleanupEnv() {
	env = ""
}

func TestFormatError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "prod"
	s := formatError(errors.New("test error"))
	assert.Equal(t, "", s)

	env = "dev"
	s = formatError(errors.New("test error"))
	lines := strings.Split(s, "\n")
	assert.Equal(t, "[DEV] test error", lines[0])
}

func TestHandleErrOnBadRequest(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	f := func(rw http.ResponseWriter, req *http.Request) {
		handleErr(req.Context(), rw, errors.New("test error"), http.StatusBadRequest)
	}
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()

	f(rec, req)

	resp := rec.Result()

	errResponse := ErrorResponse{}
	err = json.NewDecoder(resp.Body).Decode(&errResponse)
	if err != nil {
		t.Fatal(err)
	}

	if len(errResponse.Errors) != 1 {
		t.Fatal("expecting exactly one error")
	}

	assert.Equal(t, http.StatusBadRequest, errResponse.Errors[0].Extensions.Status)
	assert.Equal(t, http.StatusText(http.StatusBadRequest), errResponse.Errors[0].Extensions.HTTPError)
	assert.Equal(t, "test error", errResponse.Errors[0].Message)
	assert.Contains(t, errResponse.Errors[0].Extensions.Stack, "[DEV] test error")
}

func TestHandleErrOnInternalServerError(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	f := func(rw http.ResponseWriter, req *http.Request) {
		handleErr(req.Context(), rw, errors.New("test error"), http.StatusInternalServerError)
	}
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()

	f(rec, req)

	resp := rec.Result()

	errResponse := ErrorResponse{}
	err = json.NewDecoder(resp.Body).Decode(&errResponse)
	if err != nil {
		t.Fatal(err)
	}

	if len(errResponse.Errors) != 1 {
		t.Fatal("expecting exactly one error")
	}
	assert.Equal(t, http.StatusInternalServerError, errResponse.Errors[0].Extensions.Status)
	assert.Equal(t, http.StatusText(http.StatusInternalServerError), errResponse.Errors[0].Extensions.HTTPError)
	assert.Equal(t, "test error", errResponse.Errors[0].Message)
	assert.Contains(t, errResponse.Errors[0].Extensions.Stack, "[DEV] test error")
}

func TestHandleErrOnNotFound(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	f := func(rw http.ResponseWriter, req *http.Request) {
		handleErr(req.Context(), rw, errors.New("test error"), http.StatusNotFound)
	}
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()

	f(rec, req)

	resp := rec.Result()

	errResponse := ErrorResponse{}
	err = json.NewDecoder(resp.Body).Decode(&errResponse)
	if err != nil {
		t.Fatal(err)
	}

	if len(errResponse.Errors) != 1 {
		t.Fatal("expecting exactly one error")
	}

	assert.Equal(t, http.StatusNotFound, errResponse.Errors[0].Extensions.Status)
	assert.Equal(t, http.StatusText(http.StatusNotFound), errResponse.Errors[0].Extensions.HTTPError)
	assert.Equal(t, "test error", errResponse.Errors[0].Message)
	assert.Contains(t, errResponse.Errors[0].Extensions.Stack, "[DEV] test error")
}

func TestHandleErrOnDefault(t *testing.T) {
	t.Cleanup(CleanupEnv)
	env = "dev"
	f := func(rw http.ResponseWriter, req *http.Request) {
		handleErr(req.Context(), rw, errors.New("unauthorized"), http.StatusUnauthorized)
	}
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()

	f(rec, req)

	resp := rec.Result()

	errResponse := ErrorResponse{}
	err = json.NewDecoder(resp.Body).Decode(&errResponse)
	if err != nil {
		t.Fatal(err)
	}

	if len(errResponse.Errors) != 1 {
		t.Fatal("expecting exactly one error")
	}

	assert.Equal(t, http.StatusUnauthorized, errResponse.Errors[0].Extensions.Status)
	assert.Equal(t, http.StatusText(http.StatusUnauthorized), errResponse.Errors[0].Extensions.HTTPError)
	assert.Equal(t, "unauthorized", errResponse.Errors[0].Message)
	assert.Contains(t, errResponse.Errors[0].Extensions.Stack, "[DEV] unauthorized")
}

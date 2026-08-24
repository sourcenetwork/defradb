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
	"io"
	"net/http"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
)

const (
	// txHeaderName is the name of the transaction header.
	// This header should contain a valid transaction id.
	txHeaderName = "x-defradb-tx"
)

type contextKey string

var (
	// txsContextKey is the context key for the transaction cache.
	txsContextKey = contextKey("txs")
	// dbContextKey is the context key for the client.TxnStore
	dbContextKey = contextKey("db")
	// colContextKey is the context key for the client.Collection
	//
	// If a transaction exists, all operations will be executed
	// in the current transaction context.
	colContextKey = contextKey("col")
	// ctxContextKey is the context key for the server context.
	ctxContextKey = contextKey("ctx")
	// nodeOptsContextKey is the context key for the node options.
	nodeOptsContextKey = contextKey("nodeOpts")
)

// mustGetContextClientCollection returns the client collection from the http request context or panics.
//
// This should only be called from functions within the http package.
func mustGetContextClientCollection(req *http.Request) client.Collection {
	return req.Context().Value(colContextKey).(client.Collection) //nolint:forcetypeassert
}

// mustGetContextTxnCache returns the transaction cache from the http request context or panics.
//
// This should only be called from functions within the http package.
func mustGetContextTxnCache(req *http.Request) *txnCache {
	return req.Context().Value(txsContextKey).(*txnCache) //nolint:forcetypeassert
}

// mustGetContextClientDB returns the DB from the http request context or panics.
//
// This should only be called from functions within the http package.
func mustGetContextClientDB(req *http.Request) DB {
	return req.Context().Value(dbContextKey).(DB) //nolint:forcetypeassert
}

// tryGetContextNodeOptions returns the node options from the http request context, or nil if absent.
func tryGetContextNodeOptions(req *http.Request) *options.NodeOptions {
	opts, _ := req.Context().Value(nodeOptsContextKey).(*options.NodeOptions)
	return opts
}

// isDevMode reports whether the node serving this request has development mode enabled.
func isDevMode(req *http.Request) bool {
	opts := tryGetContextNodeOptions(req)
	return opts != nil && opts.EnableDevelopment
}

// tryGetContexCtx returns the server context if it exists.
//
// This should only be called from functions within the http package.
func tryGetContexCtx(req *http.Request) (context.Context, bool) {
	ctx, ok := req.Context().Value(ctxContextKey).(context.Context)
	return ctx, ok
}

func requestJSON(req *http.Request, out any) error {
	data, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

// requestJSONPreserveNumbers behaves like requestJSON, but decodes numbers as json.Number
// instead of a lossy float64. Needed for endpoints that carry untyped, arbitrary-precision
// numeric values (e.g. Lens module Arguments, a map[string]any) where the default float64
// decode can silently round large integers before they're ever stored.
func requestJSONPreserveNumbers(req *http.Request, out any) error {
	data, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	return decoder.Decode(out)
}

// responseJSON writes a json response with the given status and data
// to the response writer. Any errors encountered will be logged.
func responseJSON(rw http.ResponseWriter, status int, data any) {
	rw.Header().Add("Content-Type", "application/json")
	rw.WriteHeader(status)

	err := json.NewEncoder(rw).Encode(data)
	if err != nil {
		log.ErrorE("failed to write response", err)
	}
}

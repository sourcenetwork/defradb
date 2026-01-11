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
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/sourcenetwork/defradb/client"
)

const (
	// txHeaderName is the name of the transaction header.
	// This header should contain a valid transaction id.
	txHeaderName = "x-defradb-tx"
)

type contextKey string

var (
	// txsContextKey is the context key for the transaction *sync.Map
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
)

// mustGetContextClientCollection returns the client collection from the http request context or panics.
//
// This should only be called from functions within the http package.
func mustGetContextClientCollection(req *http.Request) client.Collection {
	return req.Context().Value(colContextKey).(client.Collection) //nolint:forcetypeassert
}

// mustGetContextSyncMap returns the sync map from the http request context or panics.
//
// This should only be called from functions within the http package.
func mustGetContextSyncMap(req *http.Request) *sync.Map {
	return req.Context().Value(txsContextKey).(*sync.Map) //nolint:forcetypeassert
}

// mustGetContextClientDB returns the DB from the http request context or panics.
//
// This should only be called from functions within the http package.
func mustGetContextClientDB(req *http.Request) DB {
	return req.Context().Value(dbContextKey).(DB) //nolint:forcetypeassert
}

// mustGetDataStoreTxn returns the datastore transaction or panics.
//
// This should only be called from functions within the http package.
func mustGetDataStoreTxn(tx any) client.Txn {
	return tx.(client.Txn) //nolint:forcetypeassert
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

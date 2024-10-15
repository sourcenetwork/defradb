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
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"golang.org/x/exp/slices"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/db"
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
	// dbContextKey is the context key for the client.DB
	dbContextKey = contextKey("db")
	// colContextKey is the context key for the client.Collection
	//
	// If a transaction exists, all operations will be executed
	// in the current transaction context.
	colContextKey = contextKey("col")
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

// mustGetContextClientDB returns the client DB from the http request context or panics.
//
// This should only be called from functions within the http package.
func mustGetContextClientDB(req *http.Request) client.DB {
	return req.Context().Value(dbContextKey).(client.DB) //nolint:forcetypeassert
}

// mustGetContextClientStore returns the client store from the http request context or panics.
//
// This should only be called from functions within the http package.
func mustGetContextClientStore(req *http.Request) client.Store {
	return req.Context().Value(dbContextKey).(client.Store) //nolint:forcetypeassert
}

// tryGetContextClientP2P returns the P2P client from the http request context and a boolean
// indicating if p2p was enabled.
//
// This should only be called from functions within the http package.
func tryGetContextClientP2P(req *http.Request) (client.P2P, bool) {
	p2p, ok := req.Context().Value(dbContextKey).(client.P2P)
	return p2p, ok
}

// CorsMiddleware handles cross origin request
func CorsMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowOriginFunc: func(r *http.Request, origin string) bool {
			if slices.Contains(allowedOrigins, "*") {
				return true
			}
			return slices.Contains(allowedOrigins, strings.ToLower(origin))
		},
		AllowedMethods: []string{"GET", "HEAD", "POST", "PATCH", "DELETE"},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         300,
	})
}

// ApiMiddleware sets the required context values for all API requests.
func ApiMiddleware(db client.DB, txs *sync.Map) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			ctx = context.WithValue(ctx, dbContextKey, db)
			ctx = context.WithValue(ctx, txsContextKey, txs)
			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	}
}

// TransactionMiddleware sets the transaction context for the current request.
func TransactionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		txs := mustGetContextSyncMap(req)

		txValue := req.Header.Get(txHeaderName)
		if txValue == "" {
			next.ServeHTTP(rw, req)
			return
		}
		id, err := strconv.ParseUint(txValue, 10, 64)
		if err != nil {
			next.ServeHTTP(rw, req)
			return
		}
		tx, ok := txs.Load(id)
		if !ok {
			next.ServeHTTP(rw, req)
			return
		}
		ctx := req.Context()
		if val, ok := tx.(datastore.Txn); ok {
			ctx = db.SetContextTxn(ctx, val)
		}
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}

// CollectionMiddleware sets the collection context for the current request.
func CollectionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		db := mustGetContextClientDB(req)

		col, err := db.GetCollectionByName(req.Context(), chi.URLParam(req, "name"))
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		ctx := context.WithValue(req.Context(), colContextKey, col)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}

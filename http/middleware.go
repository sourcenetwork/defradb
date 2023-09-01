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
)

const TX_HEADER_NAME = "x-defradb-tx"

type contextKey string

var (
	// txsContextKey is the context key for the transaction *sync.Map
	txsContextKey = contextKey("txs")
	// dbContextKey is the context key for the client.DB
	dbContextKey = contextKey("db")
	// txContextKey is the context key for the datastore.Txn
	//
	// NOTE: this will only be set if a transaction id is specified
	txContextKey = contextKey("tx")
	// storeContextKey is the context key for the client.Store
	//
	// NOTE: if a transaction exists, all operations will be executed
	// in the current transaction context
	storeContextKey = contextKey("store")
	// lensContextKey is the context key for the client.LensRegistry
	//
	// NOTE: if a transaction exists, all operations will be executed
	// in the current transaction context
	lensContextKey = contextKey("lens")
	// colContextKey is the context key for the client.Collection
	//
	// NOTE: if a transaction exists, all operations will be executed
	// in the current transaction context
	colContextKey = contextKey("col")
)

// CorsMiddleware handles cross origin request
func CorsMiddleware(opts ServerOptions) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowOriginFunc: func(r *http.Request, origin string) bool {
			return slices.Contains[string](opts.AllowedOrigins, strings.ToLower(origin))
		},
		AllowedMethods: []string{"GET", "HEAD", "POST", "PATCH", "DELETE"},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         300,
	})
}

// ApiMiddleware sets the required context values for all API requests.
func ApiMiddleware(db client.DB, txs *sync.Map, opts ServerOptions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if opts.TLS.HasValue() {
				rw.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
			}

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
		txs := req.Context().Value(txsContextKey).(*sync.Map)

		txValue := req.Header.Get(TX_HEADER_NAME)
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

		ctx := context.WithValue(req.Context(), txContextKey, tx)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}

// StoreMiddleware sets the db context for the current request.
func StoreMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		db := req.Context().Value(dbContextKey).(client.DB)

		var store client.Store
		if tx, ok := req.Context().Value(txContextKey).(datastore.Txn); ok {
			store = db.WithTxn(tx)
		} else {
			store = db
		}

		ctx := context.WithValue(req.Context(), storeContextKey, store)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}

// LensMiddleware sets the lens context for the current request.
func LensMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		store := req.Context().Value(storeContextKey).(client.Store)

		var lens client.LensRegistry
		if tx, ok := req.Context().Value(txContextKey).(datastore.Txn); ok {
			lens = store.LensRegistry().WithTxn(tx)
		} else {
			lens = store.LensRegistry()
		}

		ctx := context.WithValue(req.Context(), lensContextKey, lens)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}

// CollectionMiddleware sets the collection context for the current request.
func CollectionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		store := req.Context().Value(storeContextKey).(client.Store)

		col, err := store.GetCollectionByName(req.Context(), chi.URLParam(req, "name"))
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		if tx, ok := req.Context().Value(txContextKey).(datastore.Txn); ok {
			col = col.WithTxn(tx)
		}

		ctx := context.WithValue(req.Context(), colContextKey, col)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}

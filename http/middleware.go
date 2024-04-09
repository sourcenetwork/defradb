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
	"github.com/sourcenetwork/defradb/db/session"
)

const TX_HEADER_NAME = "x-defradb-tx"

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

		// store transaction in session
		sess := session.New(req.Context())
		if val, ok := tx.(datastore.Txn); ok {
			sess = sess.WithTxn(val)
		}
		next.ServeHTTP(rw, req.WithContext(sess))
	})
}

// CollectionMiddleware sets the collection context for the current request.
func CollectionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		db := req.Context().Value(dbContextKey).(client.DB)

		col, err := db.GetCollectionByName(req.Context(), chi.URLParam(req, "name"))
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		ctx := context.WithValue(req.Context(), colContextKey, col)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}

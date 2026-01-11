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
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"golang.org/x/exp/slices"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/db"
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
		AllowedHeaders: []string{"Content-Type", "Authorization"},
		MaxAge:         300,
	})
}

// ApiMiddleware sets the required context values for all API requests.
func ApiMiddleware(db client.TxnStore, txs *sync.Map) func(http.Handler) http.Handler {
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

		// Determine the storage key based on user identity
		var storageKey string
		identity := acpIdentity.FromContext(req.Context())
		if identity.HasValue() {
			// Authenticated user: construct DID-scoped key
			storageKey = identity.Value().DID() + ":" + txValue
		} else {
			// Anonymous user: the token IS the storage key
			storageKey = txValue
		}

		tx, ok := txs.Load(storageKey)
		if !ok {
			next.ServeHTTP(rw, req)
			return
		}
		ctx := req.Context()
		if val, ok := tx.(client.Txn); ok {
			ctx = db.InitContext(ctx, val)
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
			if errors.Is(err, client.ErrNotAuthorizedToPerformOperation) {
				rw.WriteHeader(http.StatusUnauthorized)
				_, _ = fmt.Fprintln(rw, err.Error())
				return
			}
			rw.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprintln(rw, err.Error())
			return
		}

		ctx := context.WithValue(req.Context(), colContextKey, col)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}

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
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/datastore"

	"github.com/go-chi/chi/v5"
)

// Version is the identifier for the current API version.
var Version string = "v0"

// playgroundHandler is set when building with the playground build tag
var playgroundHandler http.Handler = http.HandlerFunc(http.NotFound)

func NewApiRouter() (*Router, error) {
	tx_handler := &txHandler{}
	store_handler := &storeHandler{}
	acp_handler := &acpHandler{}
	collection_handler := &collectionHandler{}
	p2p_handler := &p2pHandler{}
	lens_handler := &lensHandler{}
	ccip_handler := &ccipHandler{}

	router, err := NewRouter()
	if err != nil {
		return nil, err
	}

	tx_handler.bindRoutes(router)
	store_handler.bindRoutes(router)
	acp_handler.bindRoutes(router)
	p2p_handler.bindRoutes(router)
	ccip_handler.bindRoutes(router)

	router.AddRouteGroup(func(r *Router) {
		r.AddMiddleware(CollectionMiddleware)
		collection_handler.bindRoutes(r)
	})

	router.AddRouteGroup(func(r *Router) {
		lens_handler.bindRoutes(r)
	})

	if err := router.Validate(context.Background()); err != nil {
		return nil, err
	}
	return router, nil
}

type Handler struct {
	db  client.DB
	mux *chi.Mux
	txs *sync.Map
}

func NewHandler(db client.DB) (*Handler, error) {
	router, err := NewApiRouter()
	if err != nil {
		return nil, err
	}
	// generate a secure random audience value
	audienceBytes, err := crypto.RandomBytes(64)
	if err != nil {
		return nil, err
	}
	audience := hex.EncodeToString(audienceBytes)
	txs := &sync.Map{}

	mux := chi.NewMux()
	mux.Route("/api/"+Version, func(r chi.Router) {
		r.Use(
			ApiMiddleware(db, txs),
			TransactionMiddleware,
			AuthMiddleware(audience),
		)
		r.Get("/audience", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			// ignore errors because the headers have
			// already been written to the response stream
			w.Write([]byte(audience)) //nolint:errcheck
		})
		r.Handle("/*", router)
	})
	mux.Get("/openapi.json", func(rw http.ResponseWriter, req *http.Request) {
		responseJSON(rw, http.StatusOK, router.OpenAPI())
	})
	mux.Handle("/*", playgroundHandler)

	return &Handler{
		db:  db,
		mux: mux,
		txs: txs,
	}, nil
}

func (h *Handler) Transaction(id uint64) (datastore.Txn, error) {
	tx, ok := h.txs.Load(id)
	if !ok {
		return nil, fmt.Errorf("invalid transaction id")
	}
	return tx.(datastore.Txn), nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.mux.ServeHTTP(w, req)
}

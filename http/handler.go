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
	"sync"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/event"

	"github.com/go-chi/chi/v5"
)

// Global variable for the development mode flag
// This is checked by the http/handler_extras.go/Purge function to determine which response to send
var IsDevMode bool = false

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
	extras_handler := &extrasHandler{}
	block_handler := &blockHandler{}

	router, err := NewRouter()
	if err != nil {
		return nil, err
	}

	tx_handler.bindRoutes(router)
	store_handler.bindRoutes(router)
	acp_handler.bindRoutes(router)
	p2p_handler.bindRoutes(router)
	ccip_handler.bindRoutes(router)
	extras_handler.bindRoutes(router)
	block_handler.bindRoutes(router)

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

type DB interface {
	client.DB
	// Events returns the database event queue.
	//
	// It may be used to monitor database events - a new event will be yielded for each mutation.
	// Note: it does not copy the queue, just the reference to it.
	Events() *event.Bus
}

type Handler struct {
	db  DB
	mux *chi.Mux
	txs *sync.Map
}

func NewHandler(db DB, p2p client.P2P) (*Handler, error) {
	router, err := NewApiRouter()
	if err != nil {
		return nil, err
	}
	txs := &sync.Map{}
	mux := chi.NewMux()
	mux.Route("/api/"+Version, func(r chi.Router) {
		r.Use(
			ApiMiddleware(db, p2p, txs),
			TransactionMiddleware,
			AuthMiddleware,
		)
		r.Handle("/*", router)
	})
	mux.Get("/openapi.json", func(rw http.ResponseWriter, req *http.Request) {
		responseJSON(rw, http.StatusOK, router.OpenAPI())
	})
	mux.Get("/health-check", func(rw http.ResponseWriter, req *http.Request) {
		responseJSON(rw, http.StatusOK, "Healthy")
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
		return nil, ErrInvalidTransactionId
	}

	return mustGetDataStoreTxn(tx), nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.mux.ServeHTTP(w, req)
}

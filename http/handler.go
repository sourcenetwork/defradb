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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/event"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	// VersionV0 is the identifier for the v0 API version.
	//
	// This is left in for backwards compatibility as we
	// transition to v1 and should be removed in v2.
	VersionV0 string = "v0"
	// Version is the identifier for the v1 API version.
	Version string = "v1"
)

// explorerHandler is set when building with the explorer build tag
var explorerHandler http.Handler = http.HandlerFunc(http.NotFound)

func NewApiRouter() (*Router, error) {
	tx_handler := &txHandler{}
	store_handler := &storeHandler{}
	acp_handler := &acpHandler{}
	collection_handler := &collectionHandler{}
	p2p_handler := &p2pHandler{}
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

	if err := router.Validate(context.Background()); err != nil {
		return nil, err
	}
	return router, nil
}

type DB interface {
	client.TxnStore
	client.P2P
	// Events returns the database event queue.
	//
	// It may be used to monitor database events - a new event will be yielded for each mutation.
	// Note: it does not copy the queue, just the reference to it.
	Events() event.Bus
}

type Handler struct {
	mux       *chi.Mux
	txs       *txnCache
	ctxCancel context.CancelFunc
}

func NewHandler(db DB, nodeOpts *options.NodeOptions) (*Handler, error) {
	router, err := NewApiRouter()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	txs, err := newTxnCache(ctx, &nodeOpts.HTTP)
	if err != nil {
		cancel()
		return nil, err
	}

	mux := chi.NewMux()
	// Normalize trailing slashes so that, for example, `/collections` and
	// `/collections/` resolve to the same route instead of the latter missing
	// chi's router and returning a bare 404. StripSlashes rewrites the routing
	// path in place (no redirect), preserving the request method and body.
	// It is registered before any routes so every HTTP route on this handler
	// is normalized consistently.
	mux.Use(middleware.StripSlashes)
	var allowedOrigins []string
	if nodeOpts != nil {
		allowedOrigins = nodeOpts.HTTP.AllowedOrigins
	}
	mux.Route("/api", func(r chi.Router) {
		r.Use(
			ApiMiddleware(db, txs, nodeOpts),
			TransactionMiddleware,
			AuthMiddleware(allowedOrigins),
		)
		// This is left in for backwards compatibility as we
		// transition to v1 and should be removed in v2.
		r.Mount("/"+VersionV0, router)
		r.Mount("/"+Version, router)
		r.Handle("/*", router)
	})
	mux.Get("/openapi.json", func(rw http.ResponseWriter, req *http.Request) {
		responseJSON(rw, http.StatusOK, router.OpenAPI())
	})
	mux.Get("/health-check", func(rw http.ResponseWriter, req *http.Request) {
		responseJSON(rw, http.StatusOK, "Healthy")
	})
	mux.Handle("/*", explorerHandler)
	return &Handler{
		mux:       mux,
		txs:       txs,
		ctxCancel: cancel,
	}, nil
}

func (h *Handler) Transaction(id uint64) (client.Txn, error) {
	tx, ok := h.txs.Load(id)
	if !ok {
		return nil, ErrInvalidTransactionId
	}

	return tx, nil
}

// Close stops background handler resources.
func (h *Handler) Close() {
	h.ctxCancel()
	h.txs.Close()
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.mux.ServeHTTP(w, req)
}

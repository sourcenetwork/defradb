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
	"time"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/ttl"

	"github.com/go-chi/chi/v5"
)

type TxnTTLCache = ttl.Cache[uint64, client.Txn]

// IsDevMode is a global variable for the development mode flag
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

type HandlerOptions struct {
	// Resolution of the transaction TTL
	TxnTTLTick time.Duration
	// Number of buckets for the transaction TTL cache
	TxnTTLBuckets int
}

// DefaultOpts returns the default options for the server.
func DefaultHandlerOptions() *HandlerOptions {
	return &HandlerOptions{
		TxnTTLTick:    time.Second,
		TxnTTLBuckets: 60,
	}
}

// ServerOpt is a function that configures server options.
type HandlerOpt func(*HandlerOptions)

// WithHTTPTxnTTLCache
func WithHTTPTxnTTLCache(tick time.Duration, buckets int) HandlerOpt {
	return func(opts *HandlerOptions) {
		opts.TxnTTLTick = tick
		opts.TxnTTLBuckets = buckets
	}
}

type Handler struct {
	mux *chi.Mux
	txs *TxnTTLCache
}

func NewHandler(db DB, opts ...HandlerOpt) (*Handler, error) {
	options := DefaultHandlerOptions()
	for _, opt := range opts {
		opt(options)
	}

	router, err := NewApiRouter()
	if err != nil {
		return nil, err
	}

	// transaction ttl cache that will automatically discard after
	// the ttl has expired. This is configured with 1 second
	// resolution, with 60 "buckets" (See the ttl.Wheel type).
	// Transactions that have already been commited/discard are a
	// no-op.
	txs := ttl.NewTTLCache(
		options.TxnTTLTick,
		options.TxnTTLBuckets,
		func(txid uint64, txn client.Txn) {
			txn.Discard()
		})

	mux := chi.NewMux()
	mux.Route("/api/"+Version, func(r chi.Router) {
		r.Use(
			ApiMiddleware(db, txs),
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
		mux: mux,
		txs: txs,
	}, nil
}

func (h *Handler) Transaction(id uint64) (client.Txn, error) {
	tx, ok := h.txs.Load(id)
	if !ok {
		return nil, ErrInvalidTransactionId
	}

	return tx, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.mux.ServeHTTP(w, req)
}

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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/sourcenetwork/defradb/client"
)

type H map[string]any

type Server struct {
	db     client.DB
	router *chi.Mux
	txs    *sync.Map
}

func NewServer(db client.DB) *Server {
	txs := &sync.Map{}

	txHandler := &TxHandler{txs}
	storeHandler := &StoreHandler{}
	collectionHandler := &CollectionHandler{}
	lensHandler := &LensHandler{}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	apiMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			ctx = context.WithValue(ctx, dbContextKey, db)
			ctx = context.WithValue(ctx, txsContextKey, txs)
			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	}

	router.Route("/api/v0", func(api chi.Router) {
		api.Use(apiMiddleware, TransactionMiddleware, StoreMiddleware)
		api.Route("/tx", func(tx chi.Router) {
			tx.Post("/", txHandler.NewTxn)
			tx.Post("/concurrent", txHandler.NewConcurrentTxn)
			tx.Post("/{id}", txHandler.Commit)
			tx.Delete("/{id}", txHandler.Discard)
		})
		api.Route("/backup", func(backup chi.Router) {
			backup.Post("/export", storeHandler.BasicExport)
			backup.Post("/import", storeHandler.BasicImport)
		})
		api.Route("/schema", func(schema chi.Router) {
			schema.Post("/", storeHandler.AddSchema)
			schema.Patch("/", storeHandler.PatchSchema)
		})
		api.Route("/collections", func(collections chi.Router) {
			collections.Get("/", storeHandler.GetCollection)
			// with collection middleware
			collections_tx := collections.With(CollectionMiddleware)
			collections_tx.Get("/{name}", collectionHandler.GetAllDocKeys)
			collections_tx.Post("/{name}", collectionHandler.Create)
			collections_tx.Patch("/{name}", collectionHandler.UpdateWith)
			collections_tx.Delete("/{name}", collectionHandler.DeleteWith)
			collections_tx.Post("/{name}/indexes", collectionHandler.CreateIndex)
			collections_tx.Get("/{name}/indexes", collectionHandler.GetIndexes)
			collections_tx.Delete("/{name}/indexes/{index}", collectionHandler.DropIndex)
			collections_tx.Get("/{name}/{key}", collectionHandler.Get)
			collections_tx.Post("/{name}/{key}", collectionHandler.Save)
			collections_tx.Patch("/{name}/{key}", collectionHandler.Update)
			collections_tx.Delete("/{name}/{key}", collectionHandler.Delete)
		})
		api.Route("/lens", func(lens chi.Router) {
			lens.Use(LensMiddleware)
			lens.Get("/", lensHandler.Config)
			lens.Post("/", lensHandler.SetMigration)
			lens.Post("/reload", lensHandler.ReloadLenses)
			lens.Get("/{version}", lensHandler.HasMigration)
			lens.Post("/{version}/up", lensHandler.MigrateUp)
			lens.Post("/{version}/down", lensHandler.MigrateDown)
		})
		api.Route("/graphql", func(graphQL chi.Router) {
			graphQL.Get("/", storeHandler.ExecRequest)
			graphQL.Post("/", storeHandler.ExecRequest)
		})
		api.Route("/p2p", func(p2p chi.Router) {
			p2p.Route("/replicators", func(p2p_replicators chi.Router) {
				p2p_replicators.Get("/", storeHandler.GetAllReplicators)
				p2p_replicators.Post("/", storeHandler.SetReplicator)
				p2p_replicators.Delete("/", storeHandler.DeleteReplicator)
			})
			p2p.Route("/collections", func(p2p_collections chi.Router) {
				p2p_collections.Get("/", storeHandler.GetAllP2PCollections)
				p2p_collections.Post("/{id}", storeHandler.AddP2PCollection)
				p2p_collections.Delete("/{id}", storeHandler.RemoveP2PCollection)
			})
		})
	})

	return &Server{
		db:     db,
		router: router,
		txs:    txs,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}

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
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/sourcenetwork/defradb/client"
)

type Server struct {
	db     client.DB
	router *chi.Mux
	txs    *sync.Map
}

func NewServer(db client.DB) *Server {
	txs := &sync.Map{}

	tx_handler := &txHandler{}
	store_handler := &storeHandler{}
	collection_handler := &collectionHandler{}
	lens_handler := &lensHandler{}
	ccip_handler := &ccipHandler{}

	router := chi.NewRouter()
	router.Use(middleware.RequestLogger(&logFormatter{}))
	router.Use(middleware.Recoverer)

	router.Route("/api/v0", func(api chi.Router) {
		api.Use(ApiMiddleware(db, txs), TransactionMiddleware, StoreMiddleware)
		api.Route("/tx", func(tx chi.Router) {
			tx.Post("/", tx_handler.NewTxn)
			tx.Post("/concurrent", tx_handler.NewConcurrentTxn)
			tx.Post("/{id}", tx_handler.Commit)
			tx.Delete("/{id}", tx_handler.Discard)
		})
		api.Route("/backup", func(backup chi.Router) {
			backup.Post("/export", store_handler.BasicExport)
			backup.Post("/import", store_handler.BasicImport)
		})
		api.Route("/schema", func(schema chi.Router) {
			schema.Post("/", store_handler.AddSchema)
			schema.Patch("/", store_handler.PatchSchema)
		})
		api.Route("/collections", func(collections chi.Router) {
			collections.Get("/", store_handler.GetCollection)
			// with collection middleware
			collections_tx := collections.With(CollectionMiddleware)
			collections_tx.Get("/{name}", collection_handler.GetAllDocKeys)
			collections_tx.Post("/{name}", collection_handler.Create)
			collections_tx.Patch("/{name}", collection_handler.UpdateWith)
			collections_tx.Delete("/{name}", collection_handler.DeleteWith)
			collections_tx.Post("/{name}/indexes", collection_handler.CreateIndex)
			collections_tx.Get("/{name}/indexes", collection_handler.GetIndexes)
			collections_tx.Delete("/{name}/indexes/{index}", collection_handler.DropIndex)
			collections_tx.Get("/{name}/{key}", collection_handler.Get)
			collections_tx.Patch("/{name}/{key}", collection_handler.Update)
			collections_tx.Delete("/{name}/{key}", collection_handler.Delete)
		})
		api.Route("/lens", func(lens chi.Router) {
			lens.Use(LensMiddleware)
			lens.Get("/", lens_handler.Config)
			lens.Post("/", lens_handler.SetMigration)
			lens.Post("/reload", lens_handler.ReloadLenses)
			lens.Get("/{version}", lens_handler.HasMigration)
			lens.Post("/{version}/up", lens_handler.MigrateUp)
			lens.Post("/{version}/down", lens_handler.MigrateDown)
		})
		api.Route("/graphql", func(graphQL chi.Router) {
			graphQL.Get("/", store_handler.ExecRequest)
			graphQL.Post("/", store_handler.ExecRequest)
		})
		api.Route("/ccip", func(ccip chi.Router) {
			ccip.Get("/{sender}/{data}", ccip_handler.ExecCCIP)
			ccip.Post("/", ccip_handler.ExecCCIP)
		})
		api.Route("/p2p", func(p2p chi.Router) {
			p2p.Route("/replicators", func(p2p_replicators chi.Router) {
				p2p_replicators.Get("/", store_handler.GetAllReplicators)
				p2p_replicators.Post("/", store_handler.SetReplicator)
				p2p_replicators.Delete("/", store_handler.DeleteReplicator)
			})
			p2p.Route("/collections", func(p2p_collections chi.Router) {
				p2p_collections.Get("/", store_handler.GetAllP2PCollections)
				p2p_collections.Post("/{id}", store_handler.AddP2PCollection)
				p2p_collections.Delete("/{id}", store_handler.RemoveP2PCollection)
			})
		})
		api.Route("/debug", func(debug chi.Router) {
			debug.Get("/dump", store_handler.PrintDump)
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

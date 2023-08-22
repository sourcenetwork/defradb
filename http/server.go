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

	"github.com/gin-gonic/gin"

	"github.com/sourcenetwork/defradb/client"
)

type Server struct {
	db     client.DB
	router *gin.Engine
	txs    *sync.Map
}

func NewServer(db client.DB) *Server {
	txs := &sync.Map{}

	txHandler := &TxHandler{txs}
	storeHandler := &StoreHandler{}
	collectionHandler := &CollectionHandler{}
	lensHandler := &LensHandler{}

	router := gin.Default()

	api := router.Group("/api/v0")
	api.Use(TransactionMiddleware(db, txs), DatabaseMiddleware(db))

	tx := api.Group("/tx")
	tx.POST("/", txHandler.NewTxn)
	tx.POST("/concurrent", txHandler.NewConcurrentTxn)
	tx.POST("/:id", txHandler.Commit)
	tx.DELETE("/:id", txHandler.Discard)

	backup := api.Group("/backup")
	backup.POST("/export", storeHandler.BasicExport)
	backup.POST("/import", storeHandler.BasicImport)

	schema := api.Group("/schema")
	schema.POST("/", storeHandler.AddSchema)
	schema.PATCH("/", storeHandler.PatchSchema)

	collections := api.Group("/collections")
	collections.GET("/", storeHandler.GetCollection)

	collections_tx := collections.Group("/")
	collections_tx.Use(CollectionMiddleware())

	collections_tx.GET("/:name", collectionHandler.GetAllDocKeys)
	collections_tx.POST("/:name", collectionHandler.Create)
	collections_tx.PATCH("/:name", collectionHandler.UpdateWith)
	collections_tx.DELETE("/:name", collectionHandler.DeleteWith)
	collections_tx.POST("/:name/indexes", collectionHandler.CreateIndex)
	collections_tx.GET("/:name/indexes", collectionHandler.GetIndexes)
	collections_tx.DELETE("/:name/indexes/:index", collectionHandler.DropIndex)
	collections_tx.GET("/:name/:key", collectionHandler.Get)
	collections_tx.POST("/:name/:key", collectionHandler.Save)
	collections_tx.PATCH("/:name/:key", collectionHandler.Update)
	collections_tx.DELETE("/:name/:key", collectionHandler.Delete)

	lens := api.Group("/lens")
	lens.Use(LensMiddleware())

	lens.GET("/", lensHandler.Config)
	lens.POST("/", lensHandler.SetMigration)
	lens.POST("/reload", lensHandler.ReloadLenses)
	lens.GET("/:version", lensHandler.HasMigration)
	lens.POST("/:version/up", lensHandler.MigrateUp)
	lens.POST("/:version/down", lensHandler.MigrateDown)

	graphQL := api.Group("/graphql")
	graphQL.GET("/", storeHandler.ExecRequest)
	graphQL.POST("/", storeHandler.ExecRequest)

	p2p := api.Group("/p2p")
	p2p_replicators := p2p.Group("/replicators")
	p2p_replicators.GET("/", storeHandler.GetAllReplicators)
	p2p_replicators.POST("/", storeHandler.SetReplicator)
	p2p_replicators.DELETE("/", storeHandler.DeleteReplicator)

	p2p_collections := p2p.Group("/collections")
	p2p_collections.GET("/", storeHandler.GetAllP2PCollections)
	p2p_collections.POST("/:id", storeHandler.AddP2PCollection)
	p2p_collections.DELETE("/:id", storeHandler.RemoveP2PCollection)

	return &Server{
		db:     db,
		router: router,
		txs:    txs,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}

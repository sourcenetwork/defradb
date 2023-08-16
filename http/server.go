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

// txHeaderName is the name of the custom
// header containing the transaction id.
const txHeaderName = "x-defradb-tx"

type Server struct {
	db     client.DB
	router *gin.Engine
}

func NewServer(db client.DB) *Server {
	txs := &sync.Map{}

	txHandler := &TxHandler{txs}
	storeHandler := &StoreHandler{}
	collectionHandler := &CollectionHandler{}
	lensHandler := &LensHandler{}

	router := gin.Default()
	api := router.Group("/api/v0")
	api.Use(DatabaseMiddleware(db), TransactionMiddleware(txs))

	tx := api.Group("/tx")
	tx.POST("/", txHandler.NewTxn)
	tx.POST("/concurrent", txHandler.NewConcurrentTxn)
	tx.POST("/:id", txHandler.Commit)
	tx.DELETE("/:id", txHandler.Discard)

	backup := api.Group("/backup")
	backup.POST("/export", StoreMiddleware(), storeHandler.BasicExport)
	backup.POST("/import", StoreMiddleware(), storeHandler.BasicImport)

	schema := api.Group("/schema")
	schema.POST("/", StoreMiddleware(), storeHandler.AddSchema)
	schema.PATCH("/", StoreMiddleware(), storeHandler.PatchSchema)

	collections := api.Group("/collections")
	collections.GET("/", StoreMiddleware(), storeHandler.GetCollection)
	collections.GET("/:name", CollectionMiddleware(), collectionHandler.GetAllDocKeys)
	collections.POST("/:name", CollectionMiddleware(), collectionHandler.Create)
	collections.PATCH("/:name", CollectionMiddleware(), collectionHandler.UpdateWith)
	collections.DELETE("/:name", CollectionMiddleware(), collectionHandler.DeleteWith)
	collections.POST("/:name/indexes", CollectionMiddleware(), collectionHandler.CreateIndex)
	collections.GET("/:name/indexes", CollectionMiddleware(), collectionHandler.GetIndexes)
	collections.DELETE("/:name/indexes/:index", CollectionMiddleware(), collectionHandler.DropIndex)
	collections.GET("/:name/:key", CollectionMiddleware(), collectionHandler.Get)
	collections.POST("/:name/:key", CollectionMiddleware(), collectionHandler.Save)
	collections.PATCH("/:name/:key", CollectionMiddleware(), collectionHandler.Update)
	collections.DELETE("/:name/:key", CollectionMiddleware(), collectionHandler.Delete)

	lens := api.Group("/lens")
	lens.GET("/", LensMiddleware(), lensHandler.Config)
	lens.POST("/", LensMiddleware(), lensHandler.SetMigration)
	lens.POST("/reload", LensMiddleware(), lensHandler.ReloadLenses)
	lens.GET("/:version", LensMiddleware(), lensHandler.HasMigration)
	lens.POST("/:version/up", LensMiddleware(), lensHandler.MigrateUp)
	lens.POST("/:version/down", LensMiddleware(), lensHandler.MigrateDown)

	graphQL := api.Group("/graphql")
	graphQL.GET("/", StoreMiddleware(), storeHandler.ExecRequest)
	graphQL.POST("/", StoreMiddleware(), storeHandler.ExecRequest)

	p2p := api.Group("/p2p")
	p2p_replicators := p2p.Group("/replicators")
	p2p_replicators.GET("/", StoreMiddleware(), storeHandler.GetAllReplicators)
	p2p_replicators.POST("/", StoreMiddleware(), storeHandler.SetReplicator)
	p2p_replicators.DELETE("/", StoreMiddleware(), storeHandler.DeleteReplicator)

	p2p_collections := p2p.Group("/collections")
	p2p_collections.GET("/", StoreMiddleware(), storeHandler.GetAllP2PCollections)
	p2p_collections.POST("/:id", StoreMiddleware(), storeHandler.AddP2PCollection)
	p2p_collections.DELETE("/:id", StoreMiddleware(), storeHandler.RemoveP2PCollection)

	return &Server{
		db:     db,
		router: router,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}

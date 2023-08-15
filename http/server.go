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

	"github.com/gin-gonic/gin"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
)

// txHeaderName is the name of the custom
// header containing the transaction id.
const txHeaderName = "x-defradb-tx"

type Server struct {
	store  client.Store
	router *gin.Engine
	txMap  map[uint64]datastore.Txn
}

func NewServer(store client.Store, middleware ...gin.HandlerFunc) *Server {
	txMap := make(map[uint64]datastore.Txn)

	storeHandler := &StoreHandler{}
	collectionHandler := &CollectionHandler{}
	lensHandler := &LensHandler{}

	router := gin.Default()
	api := router.Group("/api/v0")

	api.Use(func(c *gin.Context) {
		c.Set("store", store)
		c.Next()
	})
	api.Use(middleware...)

	backup := api.Group("/backup")
	backup.POST("/export", storeHandler.BasicExport)
	backup.POST("/import", storeHandler.BasicImport)

	schema := api.Group("/schema")
	schema.POST("/", storeHandler.AddSchema)
	schema.PATCH("/", storeHandler.PatchSchema)

	collections := api.Group("/collections")
	collections.GET("/", storeHandler.GetCollection)
	collections.POST("/:name", collectionHandler.Create)
	collections.PATCH("/:name", collectionHandler.UpdateWith)
	collections.DELETE("/:name", collectionHandler.DeleteWith)
	collections.POST("/:name/indexes", collectionHandler.CreateIndex)
	collections.GET("/:name/indexes", collectionHandler.GetIndexes)
	collections.DELETE("/:name/indexes/:index", collectionHandler.DropIndex)
	collections.GET("/:name/:key", collectionHandler.Get)
	collections.POST("/:name/:key", collectionHandler.Save)
	collections.PATCH("/:name/:key", collectionHandler.Update)
	collections.DELETE("/:name/:key", collectionHandler.Delete)

	lens := api.Group("/lens")
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
	p2p_replicators.GET("/replicators", storeHandler.GetAllReplicators)
	p2p_replicators.POST("/replicators", storeHandler.SetReplicator)
	p2p_replicators.DELETE("/replicators", storeHandler.DeleteReplicator)

	p2p_collections := p2p.Group("/collections")
	p2p_collections.GET("/collections", storeHandler.GetAllP2PCollections)
	p2p_collections.POST("/collections/:id", storeHandler.AddP2PCollection)
	p2p_collections.DELETE("/collections/:id", storeHandler.RemoveP2PCollection)

	return &Server{
		store:  store,
		router: router,
		txMap:  txMap,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}

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
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/events"
)

type GraphQLRequest struct {
	Query string `json:"query" form:"query"`
}

type GraphQLResponse struct {
	Errors []string         `json:"errors,omitempty"`
	Data   []map[string]any `json:"data"`
}

type Server struct {
	store client.Store
}

func NewServer(store client.Store) *gin.Engine {
	server := &Server{store}

	router := gin.Default()
	api := router.Group("/api/v0")

	backup := api.Group("/backup")
	backup.POST("/export", server.BasicExport)
	backup.POST("/import", server.BasicImport)

	schema := api.Group("/schema")
	schema.POST("/", server.AddSchema)
	schema.PATCH("/", server.PatchSchema)

	collections := api.Group("/collections")
	collections.GET("/", server.GetCollection)

	lens := api.Group("/lens")
	lens_migration := lens.Group("/migration")
	lens_migration.POST("/", server.SetMigration)

	graphQL := api.Group("/graphql")
	graphQL.GET("/", server.ExecRequest)
	graphQL.POST("/", server.ExecRequest)

	p2p := api.Group("/p2p")
	p2p_replicators := p2p.Group("/replicators")
	p2p_replicators.GET("/replicators", server.GetAllReplicators)
	p2p_replicators.POST("/replicators", server.SetReplicator)
	p2p_replicators.DELETE("/replicators", server.DeleteReplicator)

	p2p_collections := p2p.Group("/collections")
	p2p_collections.GET("/collections", server.GetAllP2PCollections)
	p2p_collections.POST("/collections/:id", server.AddP2PCollection)
	p2p_collections.DELETE("/collections/:id", server.RemoveP2PCollection)

	return router
}

func (s *Server) SetReplicator(c *gin.Context) {
	var req client.Replicator
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := s.store.SetReplicator(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *Server) DeleteReplicator(c *gin.Context) {
	var req client.Replicator
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := s.store.DeleteReplicator(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *Server) GetAllReplicators(c *gin.Context) {
	reps, err := s.store.GetAllReplicators(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, reps)
}

func (s *Server) AddP2PCollection(c *gin.Context) {
	err := s.store.AddP2PCollection(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *Server) RemoveP2PCollection(c *gin.Context) {
	err := s.store.RemoveP2PCollection(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *Server) GetAllP2PCollections(c *gin.Context) {
	cols, err := s.store.GetAllP2PCollections(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cols)
}

func (s *Server) BasicImport(c *gin.Context) {
	var config client.BackupConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := s.store.BasicImport(c.Request.Context(), config.Filepath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *Server) BasicExport(c *gin.Context) {
	var config client.BackupConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := s.store.BasicExport(c.Request.Context(), &config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *Server) AddSchema(c *gin.Context) {
	schema, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cols, err := s.store.AddSchema(c.Request.Context(), string(schema))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cols)
}

func (s *Server) PatchSchema(c *gin.Context) {
	patch, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = s.store.PatchSchema(c.Request.Context(), string(patch))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *Server) SetMigration(c *gin.Context) {
	var req client.LensConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := s.store.SetMigration(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *Server) GetCollection(c *gin.Context) {
	switch {
	case c.Query("name") != "":
		col, err := s.store.GetCollectionByName(c.Request.Context(), c.Query("name"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, col.Description())
	case c.Query("schema_id") != "":
		col, err := s.store.GetCollectionBySchemaID(c.Request.Context(), c.Query("schema_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, col.Description())
	case c.Query("version_id") != "":
		col, err := s.store.GetCollectionByVersionID(c.Request.Context(), c.Query("version_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, col.Description())
	default:
		cols, err := s.store.GetAllCollections(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		colDesc := make([]client.CollectionDescription, len(cols))
		for i, col := range cols {
			colDesc[i] = col.Description()
		}
		c.JSON(http.StatusOK, colDesc)
	}
}

func (s *Server) GetAllIndexes(c *gin.Context) {
	indexes, err := s.store.GetAllIndexes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, indexes)
}

func (s *Server) ExecRequest(c *gin.Context) {
	var request GraphQLRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if request.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing request"})
		return
	}
	result := s.store.ExecRequest(c.Request.Context(), request.Query)
	if result.Pub != nil {
		s.execRequestSubscription(c, result.Pub)
		return
	}

	var errors []string
	for _, err := range result.GQL.Errors {
		errors = append(errors, err.Error())
	}
	c.JSON(http.StatusOK, gin.H{"data": result.GQL.Data, "errors": errors})
}

func (s *Server) execRequestSubscription(c *gin.Context, pub *events.Publisher[events.Update]) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	c.Status(http.StatusOK)
	c.Writer.Flush()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Request.Context().Done():
			pub.Unsubscribe()
			return false
		case item, open := <-pub.Stream():
			if !open {
				return false
			}
			data, err := json.Marshal(item)
			if err != nil {
				return false
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			return true
		}
	})
}

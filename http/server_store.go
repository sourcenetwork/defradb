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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sourcenetwork/defradb/client"
)

type GraphQLRequest struct {
	Query string `json:"query" form:"query"`
}

type GraphQLResponse struct {
	Errors []string `json:"errors,omitempty"`
	Data   any      `json:"data"`
}

func (res *GraphQLResponse) UnmarshalJSON(data []byte) error {
	// decode numbers to json.Number
	dec := json.NewDecoder(bytes.NewBuffer(data))
	dec.UseNumber()

	var out map[string]any
	if err := dec.Decode(&out); err != nil {
		return err
	}

	// fix errors type to match tests
	switch t := out["errors"].(type) {
	case []any:
		var errors []string
		for _, v := range t {
			errors = append(errors, v.(string))
		}
		res.Errors = errors
	default:
		res.Errors = nil
	}

	// fix data type to match tests
	switch t := out["data"].(type) {
	case []any:
		var fixed []map[string]any
		for _, v := range t {
			fixed = append(fixed, v.(map[string]any))
		}
		res.Data = fixed
	case map[string]any:
		res.Data = t
	default:
		res.Data = []map[string]any{}
	}

	return nil
}

type StoreHandler struct{}

func (s *StoreHandler) SetReplicator(c *gin.Context) {
	store := c.MustGet("store").(client.Store)

	var req client.Replicator
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := store.SetReplicator(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *StoreHandler) DeleteReplicator(c *gin.Context) {
	store := c.MustGet("store").(client.Store)

	var req client.Replicator
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := store.DeleteReplicator(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *StoreHandler) GetAllReplicators(c *gin.Context) {
	store := c.MustGet("store").(client.Store)

	reps, err := store.GetAllReplicators(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, reps)
}

func (s *StoreHandler) AddP2PCollection(c *gin.Context) {
	store := c.MustGet("store").(client.Store)

	err := store.AddP2PCollection(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *StoreHandler) RemoveP2PCollection(c *gin.Context) {
	store := c.MustGet("store").(client.Store)

	err := store.RemoveP2PCollection(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *StoreHandler) GetAllP2PCollections(c *gin.Context) {
	store := c.MustGet("store").(client.Store)

	cols, err := store.GetAllP2PCollections(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cols)
}

func (s *StoreHandler) BasicImport(c *gin.Context) {
	store := c.MustGet("store").(client.Store)

	var config client.BackupConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := store.BasicImport(c.Request.Context(), config.Filepath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *StoreHandler) BasicExport(c *gin.Context) {
	store := c.MustGet("store").(client.Store)

	var config client.BackupConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := store.BasicExport(c.Request.Context(), &config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *StoreHandler) AddSchema(c *gin.Context) {
	store := c.MustGet("store").(client.Store)

	schema, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cols, err := store.AddSchema(c.Request.Context(), string(schema))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cols)
}

func (s *StoreHandler) PatchSchema(c *gin.Context) {
	store := c.MustGet("store").(client.Store)

	patch, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = store.PatchSchema(c.Request.Context(), string(patch))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *StoreHandler) GetCollection(c *gin.Context) {
	store := c.MustGet("store").(client.Store)

	switch {
	case c.Query("name") != "":
		col, err := store.GetCollectionByName(c.Request.Context(), c.Query("name"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, col.Description())
	case c.Query("schema_id") != "":
		col, err := store.GetCollectionBySchemaID(c.Request.Context(), c.Query("schema_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, col.Description())
	case c.Query("version_id") != "":
		col, err := store.GetCollectionByVersionID(c.Request.Context(), c.Query("version_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, col.Description())
	default:
		cols, err := store.GetAllCollections(c.Request.Context())
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

func (s *StoreHandler) GetAllIndexes(c *gin.Context) {
	store := c.MustGet("store").(client.Store)

	indexes, err := store.GetAllIndexes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, indexes)
}

func (s *StoreHandler) ExecRequest(c *gin.Context) {
	store := c.MustGet("store").(client.Store)

	var request GraphQLRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if request.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing request"})
		return
	}
	result := store.ExecRequest(c.Request.Context(), request.Query)

	var errors []string
	for _, err := range result.GQL.Errors {
		errors = append(errors, err.Error())
	}
	if result.Pub == nil {
		c.JSON(http.StatusOK, gin.H{"data": result.GQL.Data, "errors": errors})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	c.Status(http.StatusOK)
	c.Writer.Flush()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Request.Context().Done():
			return false
		case item, open := <-result.Pub.Stream():
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

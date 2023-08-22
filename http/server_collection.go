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
)

type CollectionHandler struct{}

type CollectionDeleteRequest struct {
	Key    string   `json:"key"`
	Keys   []string `json:"keys"`
	Filter any      `json:"filter"`
}

type CollectionUpdateRequest struct {
	Key     string   `json:"key"`
	Keys    []string `json:"keys"`
	Filter  any      `json:"filter"`
	Updater string   `json:"updater"`
}

func (s *CollectionHandler) Create(c *gin.Context) {
	col := c.MustGet("col").(client.Collection)

	var body any
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	switch t := body.(type) {
	case []map[string]any:
		var docList []*client.Document
		for _, docMap := range t {
			doc, err := client.NewDocFromMap(docMap)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			docList = append(docList, doc)
		}
		if err := col.CreateMany(c.Request.Context(), docList); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	case map[string]any:
		doc, err := client.NewDocFromMap(t)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := col.Create(c.Request.Context(), doc); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	c.Status(http.StatusOK)
}

func (s *CollectionHandler) Save(c *gin.Context) {
	col := c.MustGet("col").(client.Collection)

	var docMap map[string]any
	if err := c.ShouldBind(&docMap); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	doc, err := client.NewDocFromMap(docMap)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = col.Save(c.Request.Context(), doc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *CollectionHandler) DeleteWith(c *gin.Context) {
	col := c.MustGet("col").(client.Collection)

	var request CollectionDeleteRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	switch {
	case request.Filter != nil:
		result, err := col.DeleteWith(c.Request.Context(), request.Filter)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	case request.Key != "":
		docKey, err := client.NewDocKeyFromString(request.Key)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := col.DeleteWith(c.Request.Context(), docKey)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	case request.Keys != nil:
		var docKeys []client.DocKey
		for _, key := range request.Keys {
			docKey, err := client.NewDocKeyFromString(key)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			docKeys = append(docKeys, docKey)
		}
		result, err := col.DeleteWith(c.Request.Context(), docKeys)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid delete request"})
	}
}

func (s *CollectionHandler) UpdateWith(c *gin.Context) {
	col := c.MustGet("col").(client.Collection)

	var request CollectionUpdateRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	switch {
	case request.Filter != nil:
		result, err := col.UpdateWith(c.Request.Context(), request.Filter, request.Updater)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	case request.Key != "":
		docKey, err := client.NewDocKeyFromString(request.Key)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := col.UpdateWith(c.Request.Context(), docKey, request.Updater)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	case request.Keys != nil:
		var docKeys []client.DocKey
		for _, key := range request.Keys {
			docKey, err := client.NewDocKeyFromString(key)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			docKeys = append(docKeys, docKey)
		}
		result, err := col.UpdateWith(c.Request.Context(), docKeys, request.Updater)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid update request"})
	}
}

func (s *CollectionHandler) Update(c *gin.Context) {
	col := c.MustGet("col").(client.Collection)

	var docMap map[string]any
	if err := c.ShouldBindJSON(&docMap); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	doc, err := client.NewDocFromMap(docMap)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if doc.Key().String() != c.Param("key") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document key does not match"})
		return
	}
	err = col.Update(c.Request.Context(), doc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *CollectionHandler) Delete(c *gin.Context) {
	col := c.MustGet("col").(client.Collection)

	docKey, err := client.NewDocKeyFromString(c.Param("key"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err = col.Delete(c.Request.Context(), docKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *CollectionHandler) Get(c *gin.Context) {
	col := c.MustGet("col").(client.Collection)

	docKey, err := client.NewDocKeyFromString(c.Param("key"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err = col.Get(c.Request.Context(), docKey, c.Query("deleted") != "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

type DocKeyResult struct {
	Key   string `json:"key"`
	Error string `json:"error"`
}

func (s *CollectionHandler) GetAllDocKeys(c *gin.Context) {
	col := c.MustGet("col").(client.Collection)

	docKeyCh, err := col.GetAllDocKeys(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	c.Status(http.StatusOK)
	c.Writer.Flush()

	c.Stream(func(w io.Writer) bool {
		docKey, open := <-docKeyCh
		if !open {
			return false
		}
		results := &DocKeyResult{
			Key: docKey.Key.String(),
		}
		if docKey.Err != nil {
			results.Error = docKey.Err.Error()
		}
		data, err := json.Marshal(results)
		if err != nil {
			return false
		}
		fmt.Fprintf(w, "data: %s\n\n", data)
		return true
	})
}

func (s *CollectionHandler) CreateIndex(c *gin.Context) {
	col := c.MustGet("col").(client.Collection)

	var indexDesc client.IndexDescription
	if err := c.ShouldBind(&indexDesc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	index, err := col.CreateIndex(c.Request.Context(), indexDesc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, index)
}

func (s *CollectionHandler) GetIndexes(c *gin.Context) {
	col := c.MustGet("col").(client.Collection)

	indexes, err := col.GetIndexes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, indexes)
}

func (s *CollectionHandler) DropIndex(c *gin.Context) {
	col := c.MustGet("col").(client.Collection)

	err := col.DropIndex(c.Request.Context(), c.Param("index"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

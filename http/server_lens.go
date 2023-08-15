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
	"github.com/sourcenetwork/immutable/enumerable"

	"github.com/sourcenetwork/defradb/client"
)

type LensHandler struct{}

func (s *LensHandler) ReloadLenses(c *gin.Context) {
	store := c.MustGet("store").(client.Store)

	err := store.LensRegistry().ReloadLenses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *LensHandler) SetMigration(c *gin.Context) {
	store := c.MustGet("store").(client.Store)

	var req client.LensConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := store.SetMigration(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *LensHandler) MigrateUp(c *gin.Context) {
	store := c.MustGet("store").(client.Store)

	var src enumerable.Enumerable[map[string]any]
	if err := c.ShouldBind(src); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := store.LensRegistry().MigrateUp(c.Request.Context(), src, c.Param("version"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *LensHandler) MigrateDown(c *gin.Context) {
	store := c.MustGet("store").(client.Store)

	var src enumerable.Enumerable[map[string]any]
	if err := c.ShouldBind(src); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := store.LensRegistry().MigrateDown(c.Request.Context(), src, c.Param("version"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *LensHandler) Config(c *gin.Context) {
	store := c.MustGet("store").(client.Store)

	cfgs, err := store.LensRegistry().Config(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfgs)
}

func (s *LensHandler) HasMigration(c *gin.Context) {
	store := c.MustGet("store").(client.Store)

	exists, err := store.LensRegistry().HasMigration(c.Request.Context(), c.Param("version"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "migration not found"})
		return
	}
	c.Status(http.StatusOK)
}

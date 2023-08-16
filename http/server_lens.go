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
	lens := c.MustGet("lens").(client.LensRegistry)

	err := lens.ReloadLenses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *LensHandler) SetMigration(c *gin.Context) {
	lens := c.MustGet("lens").(client.LensRegistry)

	var req client.LensConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := lens.SetMigration(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *LensHandler) MigrateUp(c *gin.Context) {
	lens := c.MustGet("lens").(client.LensRegistry)

	var src enumerable.Enumerable[map[string]any]
	if err := c.ShouldBind(src); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := lens.MigrateUp(c.Request.Context(), src, c.Param("version"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *LensHandler) MigrateDown(c *gin.Context) {
	lens := c.MustGet("lens").(client.LensRegistry)

	var src enumerable.Enumerable[map[string]any]
	if err := c.ShouldBind(src); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := lens.MigrateDown(c.Request.Context(), src, c.Param("version"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *LensHandler) Config(c *gin.Context) {
	lens := c.MustGet("lens").(client.LensRegistry)

	cfgs, err := lens.Config(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfgs)
}

func (s *LensHandler) HasMigration(c *gin.Context) {
	lens := c.MustGet("lens").(client.LensRegistry)

	exists, err := lens.HasMigration(c.Request.Context(), c.Param("version"))
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

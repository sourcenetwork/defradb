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

type LensServer struct {
	store client.Store
}

func (s *LensServer) ReloadLenses(c *gin.Context) {
	err := s.store.LensRegistry().ReloadLenses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *LensServer) SetMigration(c *gin.Context) {
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

func (s *LensServer) MigrateUp(c *gin.Context) {
	var src enumerable.Enumerable[map[string]any]
	if err := c.ShouldBind(src); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := s.store.LensRegistry().MigrateUp(c.Request.Context(), src, c.Param("version"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *LensServer) MigrateDown(c *gin.Context) {
	var src enumerable.Enumerable[map[string]any]
	if err := c.ShouldBind(src); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := s.store.LensRegistry().MigrateDown(c.Request.Context(), src, c.Param("version"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *LensServer) Config(c *gin.Context) {
	cfgs, err := s.store.LensRegistry().Config(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfgs)
}

func (s *LensServer) HasMigration(c *gin.Context) {
	exists, err := s.store.LensRegistry().HasMigration(c.Request.Context(), c.Param("version"))
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

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
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
)

const (
	TX_HEADER_NAME      = "x-defradb-tx"
	COL_TX_HEADER_NAME  = "x-defradb-col-tx"
	LENS_TX_HEADER_NAME = "x-defradb-lens-tx"
)

// TransactionMiddleware sets the transaction context for the current request.
func TransactionMiddleware(db client.DB, txs *sync.Map, header string) gin.HandlerFunc {
	return func(c *gin.Context) {
		txValue := c.GetHeader(header)
		if txValue == "" {
			c.Next()
			return
		}
		id, err := strconv.ParseUint(txValue, 10, 64)
		if err != nil {
			c.Next()
			return
		}
		tx, ok := txs.Load(id)
		if !ok {
			c.Next()
			return
		}

		c.Set(header, tx)
		c.Next()
	}
}

// DatabaseMiddleware sets the db context for the current request.
func DatabaseMiddleware(db client.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tx, ok := c.Get(TX_HEADER_NAME)
		if ok {
			c.Set("store", db.WithTxn(tx.(datastore.Txn)))
		} else {
			c.Set("store", db)
		}
		c.Set("db", db)
		c.Next()
	}
}

// LensMiddleware sets the lens context for the current request.
func LensMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		store := c.MustGet("store").(client.Store)

		tx, ok := c.Get(LENS_TX_HEADER_NAME)
		if ok {
			c.Set("lens", store.LensRegistry().WithTxn(tx.(datastore.Txn)))
		} else {
			c.Set("lens", store.LensRegistry())
		}
		c.Next()
	}
}

// CollectionMiddleware sets the collection context for the current request.
func CollectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		store := c.MustGet("store").(client.Store)

		col, err := store.GetCollectionByName(c.Request.Context(), c.Param("name"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		tx, ok := c.Get(COL_TX_HEADER_NAME)
		if ok {
			c.Set("col", col.WithTxn(tx.(datastore.Txn)))
		} else {
			c.Set("col", col)
		}
		c.Next()
	}
}

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

type TxHandler struct {
	txs *sync.Map
}

type CreateTxResponse struct {
	ID uint64 `json:"id"`
}

func (h *TxHandler) NewTxn(c *gin.Context) {
	db := c.MustGet("db").(client.DB)
	readOnly, _ := strconv.ParseBool(c.Query("read_only"))

	tx, err := db.NewTxn(c.Request.Context(), readOnly)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.txs.Store(tx.ID(), tx)

	c.JSON(http.StatusOK, &CreateTxResponse{tx.ID()})
}

func (h *TxHandler) NewConcurrentTxn(c *gin.Context) {
	db := c.MustGet("db").(client.DB)
	readOnly, _ := strconv.ParseBool(c.Query("read_only"))

	tx, err := db.NewConcurrentTxn(c.Request.Context(), readOnly)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.txs.Store(tx.ID(), tx)

	c.JSON(http.StatusOK, &CreateTxResponse{tx.ID()})
}

func (h *TxHandler) Commit(c *gin.Context) {
	txId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
		return
	}
	txVal, ok := h.txs.Load(txId)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
		return
	}
	err = txVal.(datastore.Txn).Commit(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.txs.Delete(txId)
	c.Status(http.StatusOK)
}

func (h *TxHandler) Discard(c *gin.Context) {
	txId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
		return
	}
	txVal, ok := h.txs.LoadAndDelete(txId)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
		return
	}
	txVal.(datastore.Txn).Discard(c.Request.Context())
	c.Status(http.StatusOK)
}

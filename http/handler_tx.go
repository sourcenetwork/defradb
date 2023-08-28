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

	"github.com/go-chi/chi/v5"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
)

type txHandler struct{}

type CreateTxResponse struct {
	ID uint64 `json:"id"`
}

func (h *txHandler) NewTxn(rw http.ResponseWriter, req *http.Request) {
	db := req.Context().Value(dbContextKey).(client.DB)
	txs := req.Context().Value(txsContextKey).(*sync.Map)
	readOnly, _ := strconv.ParseBool(req.URL.Query().Get("read_only"))

	tx, err := db.NewTxn(req.Context(), readOnly)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err.Error()})
		return
	}
	txs.Store(tx.ID(), tx)
	responseJSON(rw, http.StatusOK, &CreateTxResponse{tx.ID()})
}

func (h *txHandler) NewConcurrentTxn(rw http.ResponseWriter, req *http.Request) {
	db := req.Context().Value(dbContextKey).(client.DB)
	txs := req.Context().Value(txsContextKey).(*sync.Map)
	readOnly, _ := strconv.ParseBool(req.URL.Query().Get("read_only"))

	tx, err := db.NewConcurrentTxn(req.Context(), readOnly)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err.Error()})
		return
	}
	txs.Store(tx.ID(), tx)
	responseJSON(rw, http.StatusOK, &CreateTxResponse{tx.ID()})
}

func (h *txHandler) Commit(rw http.ResponseWriter, req *http.Request) {
	txs := req.Context().Value(txsContextKey).(*sync.Map)

	txId, err := strconv.ParseUint(chi.URLParam(req, "id"), 10, 64)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{"invalid transaction id"})
		return
	}
	txVal, ok := txs.Load(txId)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{"invalid transaction id"})
		return
	}
	err = txVal.(datastore.Txn).Commit(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err.Error()})
		return
	}
	txs.Delete(txId)
	rw.WriteHeader(http.StatusOK)
}

func (h *txHandler) Discard(rw http.ResponseWriter, req *http.Request) {
	txs := req.Context().Value(txsContextKey).(*sync.Map)

	txId, err := strconv.ParseUint(chi.URLParam(req, "id"), 10, 64)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{"invalid transaction id"})
		return
	}
	txVal, ok := txs.LoadAndDelete(txId)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{"invalid transaction id"})
		return
	}
	txVal.(datastore.Txn).Discard(req.Context())
	rw.WriteHeader(http.StatusOK)
}

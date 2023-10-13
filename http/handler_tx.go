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

	"github.com/getkin/kin-openapi/openapi3"
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
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
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
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	txs.Store(tx.ID(), tx)
	responseJSON(rw, http.StatusOK, &CreateTxResponse{tx.ID()})
}

func (h *txHandler) Commit(rw http.ResponseWriter, req *http.Request) {
	txs := req.Context().Value(txsContextKey).(*sync.Map)

	txId, err := strconv.ParseUint(chi.URLParam(req, "id"), 10, 64)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrInvalidTransactionId})
		return
	}
	txVal, ok := txs.Load(txId)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrInvalidTransactionId})
		return
	}
	err = txVal.(datastore.Txn).Commit(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	txs.Delete(txId)
	rw.WriteHeader(http.StatusOK)
}

func (h *txHandler) Discard(rw http.ResponseWriter, req *http.Request) {
	txs := req.Context().Value(txsContextKey).(*sync.Map)

	txId, err := strconv.ParseUint(chi.URLParam(req, "id"), 10, 64)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrInvalidTransactionId})
		return
	}
	txVal, ok := txs.LoadAndDelete(txId)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrInvalidTransactionId})
		return
	}
	txVal.(datastore.Txn).Discard(req.Context())
	rw.WriteHeader(http.StatusOK)
}

func (h *txHandler) bindRoutes(router *Router) {
	errorResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/error",
	}
	successResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/success",
	}
	createTxSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/create_tx",
	}

	txnReadOnlyQueryParam := openapi3.NewQueryParameter("read_only").
		WithDescription("Read only transaction").
		WithSchema(openapi3.NewBoolSchema().WithDefault(false))

	txnCreateResponse := openapi3.NewResponse().
		WithDescription("Transaction info").
		WithJSONSchemaRef(createTxSchema)

	txnCreate := openapi3.NewOperation()
	txnCreate.OperationID = "new_transaction"
	txnCreate.AddParameter(txnReadOnlyQueryParam)
	txnCreate.AddResponse(200, txnCreateResponse)
	txnCreate.Responses["400"] = errorResponse

	txnConcurrent := openapi3.NewOperation()
	txnConcurrent.OperationID = "new_concurrent_transaction"
	txnConcurrent.AddParameter(txnReadOnlyQueryParam)
	txnConcurrent.AddResponse(200, txnCreateResponse)
	txnConcurrent.Responses["400"] = errorResponse

	txnIdPathParam := openapi3.NewPathParameter("id").
		WithRequired(true).
		WithSchema(openapi3.NewInt64Schema())

	txnCommit := openapi3.NewOperation()
	txnCommit.OperationID = "transaction_commit"
	txnCommit.AddParameter(txnIdPathParam)
	txnCommit.Responses = make(openapi3.Responses)
	txnCommit.Responses["200"] = successResponse
	txnCommit.Responses["400"] = errorResponse

	txnDiscard := openapi3.NewOperation()
	txnDiscard.OperationID = "transaction_discard"
	txnDiscard.AddParameter(txnIdPathParam)
	txnDiscard.Responses = make(openapi3.Responses)
	txnDiscard.Responses["200"] = successResponse
	txnDiscard.Responses["400"] = errorResponse

	router.AddRoute("/txn", http.MethodPost, txnCreate, h.NewTxn)
	router.AddRoute("/txn/concurrent", http.MethodPost, txnConcurrent, h.NewConcurrentTxn)
	router.AddRoute("/txn/{id}", http.MethodPost, txnCommit, h.Commit)
	router.AddRoute("/txn/{id}", http.MethodDelete, txnDiscard, h.Discard)
}

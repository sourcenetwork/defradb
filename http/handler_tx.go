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

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"

	"github.com/sourcenetwork/defradb/datastore"
)

type txHandler struct{}

type CreateTxResponse struct {
	ID uint64 `json:"id"`
}

func (h *txHandler) NewTxn(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	txs := mustGetContextSyncMap(req)
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
	db := mustGetContextClientDB(req)
	txs := mustGetContextSyncMap(req)
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
	txs := mustGetContextSyncMap(req)

	txID, err := strconv.ParseUint(chi.URLParam(req, "id"), 10, 64)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrInvalidTransactionId})
		return
	}
	txVal, ok := txs.Load(txID)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrInvalidTransactionId})
		return
	}

	dsTxn, ok := txVal.(datastore.Txn)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrInvalidDataStoreTransaction})
		return
	}

	err = dsTxn.Commit(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	txs.Delete(txID)
	rw.WriteHeader(http.StatusOK)
}

func (h *txHandler) Discard(rw http.ResponseWriter, req *http.Request) {
	txs := mustGetContextSyncMap(req)

	txID, err := strconv.ParseUint(chi.URLParam(req, "id"), 10, 64)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrInvalidTransactionId})
		return
	}
	txVal, ok := txs.LoadAndDelete(txID)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrInvalidTransactionId})
		return
	}

	dsTxn, ok := txVal.(datastore.Txn)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrInvalidDataStoreTransaction})
		return
	}

	dsTxn.Discard(req.Context())
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
	txnCreate.Description = "Create a new transaction"
	txnCreate.Tags = []string{"transaction"}
	txnCreate.AddParameter(txnReadOnlyQueryParam)
	txnCreate.AddResponse(200, txnCreateResponse)
	txnCreate.Responses.Set("400", errorResponse)

	txnConcurrent := openapi3.NewOperation()
	txnConcurrent.OperationID = "new_concurrent_transaction"
	txnConcurrent.Description = "Create a new concurrent transaction"
	txnConcurrent.Tags = []string{"transaction"}
	txnConcurrent.AddParameter(txnReadOnlyQueryParam)
	txnConcurrent.AddResponse(200, txnCreateResponse)
	txnConcurrent.Responses.Set("400", errorResponse)

	txnIdPathParam := openapi3.NewPathParameter("id").
		WithRequired(true).
		WithSchema(openapi3.NewInt64Schema())

	txnCommit := openapi3.NewOperation()
	txnCommit.OperationID = "transaction_commit"
	txnCommit.Description = "Commit a transaction"
	txnCommit.Tags = []string{"transaction"}
	txnCommit.AddParameter(txnIdPathParam)
	txnCommit.Responses = openapi3.NewResponses()
	txnCommit.Responses.Set("200", successResponse)
	txnCommit.Responses.Set("400", errorResponse)

	txnDiscard := openapi3.NewOperation()
	txnDiscard.OperationID = "transaction_discard"
	txnDiscard.Description = "Discard a transaction"
	txnDiscard.Tags = []string{"transaction"}
	txnDiscard.AddParameter(txnIdPathParam)
	txnDiscard.Responses = openapi3.NewResponses()
	txnDiscard.Responses.Set("200", successResponse)
	txnDiscard.Responses.Set("400", errorResponse)

	router.AddRoute("/tx", http.MethodPost, txnCreate, h.NewTxn)
	router.AddRoute("/tx/concurrent", http.MethodPost, txnConcurrent, h.NewConcurrentTxn)
	router.AddRoute("/tx/{id}", http.MethodPost, txnCommit, h.Commit)
	router.AddRoute("/tx/{id}", http.MethodDelete, txnDiscard, h.Discard)
}

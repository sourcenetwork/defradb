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
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"

	"github.com/sourcenetwork/defradb/client"
)

type txHandler struct{}

type CreateTxResponse struct {
	ID uint64 `json:"id"`
}

func (h *txHandler) NewTxn(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	txs := mustGetContextTxnCache(req)
	readOnly, _ := strconv.ParseBool(req.URL.Query().Get("read_only"))
	txnTTL, hasTxnTTL, err := parseTxnTTL(req)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	tx, err := db.NewTxn(readOnly)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	if hasTxnTTL {
		err = txs.StoreFor(tx, txnTTL)
	} else {
		err = txs.Store(tx)
	}
	if err != nil {
		tx.Discard()
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, &CreateTxResponse{tx.ID()})
}

func (h *txHandler) Commit(rw http.ResponseWriter, req *http.Request) {
	txs := mustGetContextTxnCache(req)

	txID, err := strconv.ParseUint(chi.URLParam(req, "id"), 10, 64)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrInvalidTransactionId})
		return
	}
	tx, txnTTL, ok := txs.LoadAndDelete(txID)
	if !ok {
		responseJSON(rw, http.StatusNotFound, errorResponse{client.ErrTransactionNotFound})
		return
	}

	err = tx.Commit()
	if err != nil {
		if storeErr := txs.StoreFor(tx, txnTTL); storeErr != nil {
			log.ErrorE("failed to restore transaction after commit error", storeErr)
			tx.Discard()
		}
		responseJSON(rw, httpStatusFromError(err), errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h *txHandler) Discard(rw http.ResponseWriter, req *http.Request) {
	txs := mustGetContextTxnCache(req)

	txID, err := strconv.ParseUint(chi.URLParam(req, "id"), 10, 64)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrInvalidTransactionId})
		return
	}
	tx, _, ok := txs.LoadAndDelete(txID)
	if !ok {
		responseJSON(rw, http.StatusNotFound, errorResponse{client.ErrTransactionNotFound})
		return
	}

	tx.Discard()

	rw.WriteHeader(http.StatusOK)
}

func parseTxnTTL(req *http.Request) (time.Duration, bool, error) {
	raw := req.URL.Query().Get("ttl")
	if raw == "" {
		return 0, false, nil
	}
	txnTTL, err := time.ParseDuration(raw)
	if err == nil {
		return txnTTL, true, nil
	}
	seconds, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false, err
	}
	return time.Duration(seconds) * time.Second, true, nil
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
	txnTTLQueryParam := openapi3.NewQueryParameter("ttl").
		WithDescription("Transaction idle TTL as a Go duration string, or seconds if no unit is provided").
		WithSchema(openapi3.NewStringSchema())

	txnCreateResponse := openapi3.NewResponse().
		WithDescription("Transaction info").
		WithJSONSchemaRef(createTxSchema)

	txnCreate := openapi3.NewOperation()
	txnCreate.OperationID = "new_transaction"
	txnCreate.Description = "Create a new transaction"
	txnCreate.Tags = []string{"transaction"}
	txnCreate.AddParameter(txnReadOnlyQueryParam)
	txnCreate.AddParameter(txnTTLQueryParam)
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
	txnCommit.OperationID = "commit_transaction"
	txnCommit.Description = "Commit a transaction"
	txnCommit.Tags = []string{"transaction"}
	txnCommit.AddParameter(txnIdPathParam)
	txnCommit.Responses = openapi3.NewResponses()
	txnCommit.Responses.Set("200", successResponse)
	txnCommit.Responses.Set("400", errorResponse)
	txnCommit.Responses.Set("404", errorResponse)
	txnCommit.Responses.Set("409", errorResponse)

	txnDiscard := openapi3.NewOperation()
	txnDiscard.OperationID = "discard_transaction"
	txnDiscard.Description = "Discard a transaction"
	txnDiscard.Tags = []string{"transaction"}
	txnDiscard.AddParameter(txnIdPathParam)
	txnDiscard.Responses = openapi3.NewResponses()
	txnDiscard.Responses.Set("200", successResponse)
	txnDiscard.Responses.Set("400", errorResponse)
	txnDiscard.Responses.Set("404", errorResponse)

	router.AddRoute("/tx", http.MethodPost, txnCreate, h.NewTxn)
	router.AddRoute("/tx/{id}", http.MethodPost, txnCommit, h.Commit)
	router.AddRoute("/tx/{id}", http.MethodDelete, txnDiscard, h.Discard)
}

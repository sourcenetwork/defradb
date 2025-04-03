// Copyright 2025 Democratized Data Foundation
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

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/sourcenetwork/defradb/errors"
)

const (
	errMissingParameter = "missing required parameter"
	blockCidParam       = "cid"
)

type blockHandler struct{}

func (h *blockHandler) bindRoutes(router *Router) {
	errorResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/error",
	}
	successResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/success",
	}

	cidQueryParam := openapi3.NewQueryParameter("cid").
		WithDescription("Content ID of the block to verify").
		WithRequired(true).
		WithSchema(openapi3.NewStringSchema())

	verifyBlock := openapi3.NewOperation()
	verifyBlock.OperationID = "verify_block"
	verifyBlock.Description = "Verify block signature"
	verifyBlock.Tags = []string{"block"}
	verifyBlock.AddParameter(cidQueryParam)
	verifyBlock.Responses = openapi3.NewResponses()
	verifyBlock.Responses.Set("200", successResponse)
	verifyBlock.Responses.Set("400", errorResponse)

	router.AddRoute("/block/verify", http.MethodGet, verifyBlock, h.verifyBlock)
}

// verifyBlock handles block signature verification requests
func (h *blockHandler) verifyBlock(w http.ResponseWriter, r *http.Request) {
	db := mustGetContextClientDB(r)
	cid := r.URL.Query().Get(blockCidParam)
	if cid == "" {
		responseJSON(w, http.StatusBadRequest, errorResponse{
			errors.New(errMissingParameter, errors.NewKV("Parameter", "cid")),
		})
		return
	}

	err := db.VerifyBlock(r.Context(), cid)
	if err != nil {
		responseJSON(w, http.StatusBadRequest, errorResponse{err})
		return
	}

	responseJSON(w, http.StatusOK, nil)
}

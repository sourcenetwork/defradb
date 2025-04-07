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
	"github.com/sourcenetwork/defradb/crypto"
)

const (
	blockCidParam  string = "cid"
	publicKeyParam string = "public-key"
	typeParam      string = "type"
)

type blockHandler struct{}

// verifyBlock handles block signature verification requests
func (h *blockHandler) verifyBlock(w http.ResponseWriter, r *http.Request) {
	db := mustGetContextClientDB(r)
	cid := r.URL.Query().Get(blockCidParam)
	if cid == "" {
		responseJSON(w, http.StatusBadRequest, errorResponse{
			NewErrMissingRequiredParameter(blockCidParam),
		})
		return
	}

	publicKey := r.URL.Query().Get(publicKeyParam)
	if publicKey == "" {
		responseJSON(w, http.StatusBadRequest, errorResponse{
			NewErrMissingRequiredParameter(publicKeyParam),
		})
		return
	}

	keyType := crypto.KeyTypeSecp256k1
	typeStr := r.URL.Query().Get(typeParam)
	if typeStr != "" {
		keyType = crypto.KeyType(typeStr)
	}

	pubKey, err := crypto.PublicKeyFromString(keyType, publicKey)
	if err != nil {
		responseJSON(w, http.StatusBadRequest, errorResponse{err})
		return
	}

	err = db.VerifySignature(r.Context(), cid, pubKey)
	if err != nil {
		responseJSON(w, http.StatusBadRequest, errorResponse{err})
		return
	}

	responseJSON(w, http.StatusOK, nil)
}

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

	router.AddRoute("/block/verify-signature", http.MethodGet, verifyBlock, h.verifyBlock)
}

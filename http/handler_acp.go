// Copyright 2024 Democratized Data Foundation
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
	"io"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/sourcenetwork/defradb/client"
)

type acpHandler struct{}

func (s *acpHandler) AddPolicy(rw http.ResponseWriter, req *http.Request) {
	db, ok := req.Context().Value(dbContextKey).(client.DB)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{NewErrFailedToGetContext("db")})
		return
	}

	policyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	addPolicyResult, err := db.AddPolicy(
		req.Context(),
		string(policyBytes),
	)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	responseJSON(rw, http.StatusOK, addPolicyResult)
}

func (h *acpHandler) bindRoutes(router *Router) {
	successResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/success",
	}
	errorResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/error",
	}

	acpAddPolicyRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithSchema(openapi3.NewStringSchema(), []string{"text/plain"}))

	acpAddPolicy := openapi3.NewOperation()
	acpAddPolicy.OperationID = "add policy"
	acpAddPolicy.Description = "Add a policy using acp system"
	acpAddPolicy.Tags = []string{"acp_policy"}
	acpAddPolicy.Responses = openapi3.NewResponses()
	acpAddPolicy.Responses.Set("200", successResponse)
	acpAddPolicy.Responses.Set("400", errorResponse)
	acpAddPolicy.RequestBody = &openapi3.RequestBodyRef{
		Value: acpAddPolicyRequest,
	}

	router.AddRoute("/acp/policy", http.MethodPost, acpAddPolicy, h.AddPolicy)
}

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

func (s *acpHandler) AddDocActorRelationship(rw http.ResponseWriter, req *http.Request) {
	db, ok := req.Context().Value(dbContextKey).(client.DB)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{NewErrFailedToGetContext("db")})
		return
	}

	var message addDocActorRelationshipRequest
	err := requestJSON(req, &message)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	addDocActorRelResult, err := db.AddDocActorRelationship(
		req.Context(),
		message.CollectionName,
		message.DocID,
		message.Relation,
		message.TargetActor,
	)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	responseJSON(rw, http.StatusOK, addDocActorRelResult)
}

func (s *acpHandler) DeleteDocActorRelationship(rw http.ResponseWriter, req *http.Request) {
	db, ok := req.Context().Value(dbContextKey).(client.DB)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{NewErrFailedToGetContext("db")})
		return
	}

	var message deleteDocActorRelationshipRequest
	err := requestJSON(req, &message)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	deleteDocActorRelResult, err := db.DeleteDocActorRelationship(
		req.Context(),
		message.CollectionName,
		message.DocID,
		message.Relation,
		message.TargetActor,
	)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	responseJSON(rw, http.StatusOK, deleteDocActorRelResult)
}

func (h *acpHandler) bindRoutes(router *Router) {
	errorResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/error",
	}

	acpPolicyAddResultSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_policy_add_result",
	}

	acpRelationshipAddRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_relationship_add_request",
	}
	acpRelationshipAddResultSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_relationship_add_result",
	}

	acpRelationshipDeleteRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_relationship_delete_request",
	}
	acpRelationshipDeleteResultSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_relationship_delete_result",
	}

	acpAddPolicyRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithSchema(openapi3.NewStringSchema(), []string{"text/plain"}))
	acpPolicyAddResult := openapi3.NewResponse().
		WithDescription("Add acp policy result").
		WithJSONSchemaRef(acpPolicyAddResultSchema)
	acpAddPolicy := openapi3.NewOperation()
	acpAddPolicy.OperationID = "add policy"
	acpAddPolicy.Description = "Add a policy using acp system"
	acpAddPolicy.Tags = []string{"acp_policy"}
	acpAddPolicy.Responses = openapi3.NewResponses()
	acpAddPolicy.AddResponse(200, acpPolicyAddResult)
	acpAddPolicy.Responses.Set("400", errorResponse)
	acpAddPolicy.RequestBody = &openapi3.RequestBodyRef{
		Value: acpAddPolicyRequest,
	}

	acpAddDocActorRelationshipRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(acpRelationshipAddRequestSchema))
	acpAddDocActorRelationshipResult := openapi3.NewResponse().
		WithDescription("Add acp relationship result").
		WithJSONSchemaRef(acpRelationshipAddResultSchema)
	acpAddDocActorRelationship := openapi3.NewOperation()
	acpAddDocActorRelationship.OperationID = "add relationship"
	acpAddDocActorRelationship.Description = "Add an actor relationship using acp system"
	acpAddDocActorRelationship.Tags = []string{"acp_relationship"}
	acpAddDocActorRelationship.Responses = openapi3.NewResponses()
	acpAddDocActorRelationship.AddResponse(200, acpAddDocActorRelationshipResult)
	acpAddDocActorRelationship.Responses.Set("400", errorResponse)
	acpAddDocActorRelationship.RequestBody = &openapi3.RequestBodyRef{
		Value: acpAddDocActorRelationshipRequest,
	}

	acpDeleteDocActorRelationshipRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(acpRelationshipDeleteRequestSchema))
	acpDeleteDocActorRelationshipResult := openapi3.NewResponse().
		WithDescription("Delete acp relationship result").
		WithJSONSchemaRef(acpRelationshipDeleteResultSchema)
	acpDeleteDocActorRelationship := openapi3.NewOperation()
	acpDeleteDocActorRelationship.OperationID = "delete relationship"
	acpDeleteDocActorRelationship.Description = "Delete an actor relationship using acp system"
	acpDeleteDocActorRelationship.Tags = []string{"acp_relationship"}
	acpDeleteDocActorRelationship.Responses = openapi3.NewResponses()
	acpDeleteDocActorRelationship.AddResponse(200, acpDeleteDocActorRelationshipResult)
	acpDeleteDocActorRelationship.Responses.Set("400", errorResponse)
	acpDeleteDocActorRelationship.RequestBody = &openapi3.RequestBodyRef{
		Value: acpDeleteDocActorRelationshipRequest,
	}

	router.AddRoute("/acp/policy", http.MethodPost, acpAddPolicy, h.AddPolicy)
	router.AddRoute("/acp/relationship", http.MethodPost, acpAddDocActorRelationship, h.AddDocActorRelationship)
	router.AddRoute("/acp/relationship", http.MethodDelete, acpDeleteDocActorRelationship, h.DeleteDocActorRelationship)
}

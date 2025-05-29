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
)

type acpHandler struct{}

func (s *acpHandler) AddPolicyWithDAC(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	policyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	addPolicyResult, err := db.AddPolicyWithDAC(
		req.Context(),
		string(policyBytes),
	)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	responseJSON(rw, http.StatusOK, addPolicyResult)
}

func (s *acpHandler) AddActorRelationshipWithDAC(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	var message addActorRelationshipWithDACRequest
	err := requestJSON(req, &message)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	addDocActorRelResult, err := db.AddActorRelationshipWithDAC(
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

func (s *acpHandler) DeleteActorRelationshipWithDAC(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	var message deleteActorRelationshipWithDACRequest
	err := requestJSON(req, &message)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	deleteDocActorRelResult, err := db.DeleteActorRelationshipWithDAC(
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

	addPolicyDACResultSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_policy_add_result",
	}

	addRelationshipDACRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_relationship_add_request",
	}
	addRelationshipDACResultSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_relationship_add_result",
	}

	deleteRelationshipDACRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_relationship_delete_request",
	}
	deleteRelationshipDACResultSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_relationship_delete_result",
	}

	addPolicyDACRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithSchema(openapi3.NewStringSchema(), []string{"text/plain"}))
	addPolicyDACResult := openapi3.NewResponse().
		WithDescription("Add document acp policy result").
		WithJSONSchemaRef(addPolicyDACResultSchema)
	addPolicyDAC := openapi3.NewOperation()
	addPolicyDAC.OperationID = "add policy"
	addPolicyDAC.Description = "Add a policy using document acp system"
	addPolicyDAC.Tags = []string{"acp_policy"}
	addPolicyDAC.Responses = openapi3.NewResponses()
	addPolicyDAC.AddResponse(200, addPolicyDACResult)
	addPolicyDAC.Responses.Set("400", errorResponse)
	addPolicyDAC.RequestBody = &openapi3.RequestBodyRef{
		Value: addPolicyDACRequest,
	}

	addActorRelationshipDACRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(addRelationshipDACRequestSchema))
	addActorRelationshipDACResult := openapi3.NewResponse().
		WithDescription("Add document acp relationship result").
		WithJSONSchemaRef(addRelationshipDACResultSchema)
	addActorRelationshipDAC := openapi3.NewOperation()
	addActorRelationshipDAC.OperationID = "add relationship"
	addActorRelationshipDAC.Description = "Add an actor relationship using document acp system"
	addActorRelationshipDAC.Tags = []string{"acp_relationship"}
	addActorRelationshipDAC.Responses = openapi3.NewResponses()
	addActorRelationshipDAC.AddResponse(200, addActorRelationshipDACResult)
	addActorRelationshipDAC.Responses.Set("400", errorResponse)
	addActorRelationshipDAC.RequestBody = &openapi3.RequestBodyRef{
		Value: addActorRelationshipDACRequest,
	}

	deleteActorRelationshipDACRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(deleteRelationshipDACRequestSchema))
	deleteActorRelationshipDACResult := openapi3.NewResponse().
		WithDescription("Delete document acp relationship result").
		WithJSONSchemaRef(deleteRelationshipDACResultSchema)
	deleteActorRelationshipDAC := openapi3.NewOperation()
	deleteActorRelationshipDAC.OperationID = "delete relationship"
	deleteActorRelationshipDAC.Description = "Delete an actor relationship using document acp system"
	deleteActorRelationshipDAC.Tags = []string{"acp_relationship"}
	deleteActorRelationshipDAC.Responses = openapi3.NewResponses()
	deleteActorRelationshipDAC.AddResponse(200, deleteActorRelationshipDACResult)
	deleteActorRelationshipDAC.Responses.Set("400", errorResponse)
	deleteActorRelationshipDAC.RequestBody = &openapi3.RequestBodyRef{
		Value: deleteActorRelationshipDACRequest,
	}

	router.AddRoute("/acp/policy", http.MethodPost, addPolicyDAC, h.AddPolicyWithDAC)
	router.AddRoute("/acp/relationship", http.MethodPost, addActorRelationshipDAC, h.AddActorRelationshipWithDAC)
	router.AddRoute("/acp/relationship", http.MethodDelete, deleteActorRelationshipDAC, h.DeleteActorRelationshipWithDAC)
}

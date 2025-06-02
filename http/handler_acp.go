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

func (s *acpHandler) AddDACPolicy(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	policyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	addPolicyResult, err := db.AddDACPolicy(
		req.Context(),
		string(policyBytes),
	)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	responseJSON(rw, http.StatusOK, addPolicyResult)
}

func (s *acpHandler) AddDACActorRelationship(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	var message addDACActorRelationshipRequest
	err := requestJSON(req, &message)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	addDocActorRelResult, err := db.AddDACActorRelationship(
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

func (s *acpHandler) DeleteDACActorRelationship(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	var message deleteDACActorRelationshipRequest
	err := requestJSON(req, &message)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	deleteDocActorRelResult, err := db.DeleteDACActorRelationship(
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

	// Note: The result types are more general and not specific to aac or dac.
	addPolicyResultSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_policy_add_result",
	}
	addRelationshipResultSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_relationship_add_result",
	}
	deleteRelationshipResultSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_relationship_delete_result",
	}

	// Note: The request types are more specific to aac or dac.
	addRelationshipDACRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_dac_relationship_add_request",
	}
	deleteRelationshipDACRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_dac_relationship_delete_request",
	}

	addPolicyDACRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithSchema(openapi3.NewStringSchema(), []string{"text/plain"}))
	addPolicyDACResult := openapi3.NewResponse().
		WithDescription("Add document acp policy result").
		WithJSONSchemaRef(addPolicyResultSchema)
	addPolicyDAC := openapi3.NewOperation()
	addPolicyDAC.OperationID = "add dac policy"
	addPolicyDAC.Description = "Add a policy using document acp system"
	addPolicyDAC.Tags = []string{"acp_dac_policy"}
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
		WithJSONSchemaRef(addRelationshipResultSchema)
	addActorRelationshipDAC := openapi3.NewOperation()
	addActorRelationshipDAC.OperationID = "add dac relationship"
	addActorRelationshipDAC.Description = "Add an actor relationship using document acp system"
	addActorRelationshipDAC.Tags = []string{"acp_dac_relationship"}
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
		WithJSONSchemaRef(deleteRelationshipResultSchema)
	deleteActorRelationshipDAC := openapi3.NewOperation()
	deleteActorRelationshipDAC.OperationID = "delete dac relationship"
	deleteActorRelationshipDAC.Description = "Delete an actor relationship using document acp system"
	deleteActorRelationshipDAC.Tags = []string{"acp_dac_relationship"}
	deleteActorRelationshipDAC.Responses = openapi3.NewResponses()
	deleteActorRelationshipDAC.AddResponse(200, deleteActorRelationshipDACResult)
	deleteActorRelationshipDAC.Responses.Set("400", errorResponse)
	deleteActorRelationshipDAC.RequestBody = &openapi3.RequestBodyRef{
		Value: deleteActorRelationshipDACRequest,
	}

	router.AddRoute("/acp/dac/policy", http.MethodPost, addPolicyDAC, h.AddDACPolicy)
	router.AddRoute(
		"/acp/dac/relationship",
		http.MethodPost,
		addActorRelationshipDAC,
		h.AddDACActorRelationship,
	)
	router.AddRoute(
		"/acp/dac/relationship",
		http.MethodDelete,
		deleteActorRelationshipDAC,
		h.DeleteDACActorRelationship,
	)
}

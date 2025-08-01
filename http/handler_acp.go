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

func (s *acpHandler) AddNACActorRelationship(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	var message addNACActorRelationshipRequest
	err := requestJSON(req, &message)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	addActorRelationshipResult, err := db.AddNACActorRelationship(
		req.Context(),
		message.Relation,
		message.TargetActor,
	)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	responseJSON(rw, http.StatusOK, addActorRelationshipResult)
}

func (s *acpHandler) DeleteNACActorRelationship(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	var message deleteNACActorRelationshipRequest
	err := requestJSON(req, &message)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	deleteActorRelationshipResult, err := db.DeleteNACActorRelationship(
		req.Context(),
		message.Relation,
		message.TargetActor,
	)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	responseJSON(rw, http.StatusOK, deleteActorRelationshipResult)
}

func (s *acpHandler) ReEnableNAC(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	err := db.ReEnableNAC(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (s *acpHandler) DisableNAC(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	err := db.DisableNAC(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (s *acpHandler) GetNACStatus(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	statusNACResult, err := db.GetNACStatus(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	responseJSON(rw, http.StatusOK, statusNACResult)
}

func (h *acpHandler) bindRoutes(router *Router) {
	successResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/success",
	}
	errorResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/error",
	}

	addPolicyResultSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_policy_add_result",
	}
	addRelationshipResultSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_relationship_add_result",
	}
	deleteRelationshipResultSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_relationship_delete_result",
	}
	statusNACResultSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_node_status_result",
	}

	addRelationshipDACRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_dac_relationship_add_request",
	}
	deleteRelationshipDACRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_dac_relationship_delete_request",
	}

	addRelationshipNACRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_node_relationship_add_request",
	}
	deleteRelationshipNACRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_node_relationship_delete_request",
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

	addActorRelationshipNACRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(addRelationshipNACRequestSchema))
	addActorRelationshipNACResult := openapi3.NewResponse().
		WithDescription("Add node acp relationship result").
		WithJSONSchemaRef(addRelationshipResultSchema)
	addActorRelationshipNAC := openapi3.NewOperation()
	addActorRelationshipNAC.OperationID = "add nac relationship"
	addActorRelationshipNAC.Description = "Add an actor relationship using node acp system"
	addActorRelationshipNAC.Tags = []string{"acp_node_relationship"}
	addActorRelationshipNAC.Responses = openapi3.NewResponses()
	addActorRelationshipNAC.AddResponse(200, addActorRelationshipNACResult)
	addActorRelationshipNAC.Responses.Set("400", errorResponse)
	addActorRelationshipNAC.RequestBody = &openapi3.RequestBodyRef{
		Value: addActorRelationshipNACRequest,
	}

	deleteActorRelationshipNACRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(deleteRelationshipNACRequestSchema))
	deleteActorRelationshipNACResult := openapi3.NewResponse().
		WithDescription("Delete node acp relationship result").
		WithJSONSchemaRef(deleteRelationshipResultSchema)
	deleteActorRelationshipNAC := openapi3.NewOperation()
	deleteActorRelationshipNAC.OperationID = "delete nac relationship"
	deleteActorRelationshipNAC.Description = "Delete an actor relationship using node acp system"
	deleteActorRelationshipNAC.Tags = []string{"acp_node_relationship"}
	deleteActorRelationshipNAC.Responses = openapi3.NewResponses()
	deleteActorRelationshipNAC.AddResponse(200, deleteActorRelationshipNACResult)
	deleteActorRelationshipNAC.Responses.Set("400", errorResponse)
	deleteActorRelationshipNAC.RequestBody = &openapi3.RequestBodyRef{
		Value: deleteActorRelationshipNACRequest,
	}

	reEnableNAC := openapi3.NewOperation()
	reEnableNAC.OperationID = "re-enable nac"
	reEnableNAC.Description = "Re-enable nac"
	reEnableNAC.Tags = []string{"acp_node_re-enable"}
	reEnableNAC.Responses = openapi3.NewResponses()
	reEnableNAC.Responses.Set("200", successResponse)
	reEnableNAC.Responses.Set("400", errorResponse)

	disableNAC := openapi3.NewOperation()
	disableNAC.OperationID = "disable nac"
	disableNAC.Description = "Disable nac"
	disableNAC.Tags = []string{"acp_node_disable"}
	disableNAC.Responses = openapi3.NewResponses()
	disableNAC.Responses.Set("200", successResponse)
	disableNAC.Responses.Set("400", errorResponse)

	statusNACResult := openapi3.NewResponse().
		WithDescription("Node acp status result").
		WithJSONSchemaRef(statusNACResultSchema)
	statusNAC := openapi3.NewOperation()
	statusNAC.OperationID = "Check status of nac"
	statusNAC.Description = "Check status of node acp system"
	statusNAC.Tags = []string{"acp_node_status"}
	statusNAC.Responses = openapi3.NewResponses()
	statusNAC.AddResponse(200, statusNACResult)
	statusNAC.Responses.Set("400", errorResponse)

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

	router.AddRoute(
		"/acp/node/relationship",
		http.MethodPost,
		addActorRelationshipNAC,
		h.AddNACActorRelationship,
	)
	router.AddRoute(
		"/acp/node/relationship",
		http.MethodDelete,
		deleteActorRelationshipNAC,
		h.DeleteNACActorRelationship,
	)
	router.AddRoute(
		"/acp/node/re-enable",
		http.MethodPost,
		reEnableNAC,
		h.ReEnableNAC,
	)
	router.AddRoute(
		"/acp/node/disable",
		http.MethodPost,
		disableNAC,
		h.DisableNAC,
	)
	router.AddRoute(
		"/acp/node/status",
		http.MethodGet,
		statusNAC,
		h.GetNACStatus,
	)
}

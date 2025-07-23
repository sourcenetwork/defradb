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

func (s *acpHandler) AddAACActorRelationship(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	var message addAACActorRelationshipRequest
	err := requestJSON(req, &message)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	addActorRelationshipResult, err := db.AddAACActorRelationship(
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

func (s *acpHandler) DeleteAACActorRelationship(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	var message deleteAACActorRelationshipRequest
	err := requestJSON(req, &message)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	deleteActorRelationshipResult, err := db.DeleteAACActorRelationship(
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

func (s *acpHandler) ReEnableAAC(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	err := db.ReEnableAAC(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (s *acpHandler) DisableAAC(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	err := db.DisableAAC(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (s *acpHandler) GetAACStatus(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	statusAACResult, err := db.GetAACStatus(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	responseJSON(rw, http.StatusOK, statusAACResult)
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
	statusAACResultSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_aac_status_result",
	}

	addRelationshipDACRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_dac_relationship_add_request",
	}
	deleteRelationshipDACRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_dac_relationship_delete_request",
	}

	addRelationshipAACRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_aac_relationship_add_request",
	}
	deleteRelationshipAACRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/acp_aac_relationship_delete_request",
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

	addActorRelationshipAACRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(addRelationshipAACRequestSchema))
	addActorRelationshipAACResult := openapi3.NewResponse().
		WithDescription("Add admin acp relationship result").
		WithJSONSchemaRef(addRelationshipResultSchema)
	addActorRelationshipAAC := openapi3.NewOperation()
	addActorRelationshipAAC.OperationID = "add aac relationship"
	addActorRelationshipAAC.Description = "Add an actor relationship using admin acp system"
	addActorRelationshipAAC.Tags = []string{"acp_aac_relationship"}
	addActorRelationshipAAC.Responses = openapi3.NewResponses()
	addActorRelationshipAAC.AddResponse(200, addActorRelationshipAACResult)
	addActorRelationshipAAC.Responses.Set("400", errorResponse)
	addActorRelationshipAAC.RequestBody = &openapi3.RequestBodyRef{
		Value: addActorRelationshipAACRequest,
	}

	deleteActorRelationshipAACRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(deleteRelationshipAACRequestSchema))
	deleteActorRelationshipAACResult := openapi3.NewResponse().
		WithDescription("Delete admin acp relationship result").
		WithJSONSchemaRef(deleteRelationshipResultSchema)
	deleteActorRelationshipAAC := openapi3.NewOperation()
	deleteActorRelationshipAAC.OperationID = "delete aac relationship"
	deleteActorRelationshipAAC.Description = "Delete an actor relationship using admin acp system"
	deleteActorRelationshipAAC.Tags = []string{"acp_aac_relationship"}
	deleteActorRelationshipAAC.Responses = openapi3.NewResponses()
	deleteActorRelationshipAAC.AddResponse(200, deleteActorRelationshipAACResult)
	deleteActorRelationshipAAC.Responses.Set("400", errorResponse)
	deleteActorRelationshipAAC.RequestBody = &openapi3.RequestBodyRef{
		Value: deleteActorRelationshipAACRequest,
	}

	reEnableAAC := openapi3.NewOperation()
	reEnableAAC.OperationID = "re-enable aac"
	reEnableAAC.Description = "Re-enable aac"
	reEnableAAC.Tags = []string{"acp_aac_re-enable"}
	reEnableAAC.Responses = openapi3.NewResponses()
	reEnableAAC.Responses.Set("200", successResponse)
	reEnableAAC.Responses.Set("400", errorResponse)

	disableAAC := openapi3.NewOperation()
	disableAAC.OperationID = "disable aac"
	disableAAC.Description = "Disable aac"
	disableAAC.Tags = []string{"acp_aac_disable"}
	disableAAC.Responses = openapi3.NewResponses()
	disableAAC.Responses.Set("200", successResponse)
	disableAAC.Responses.Set("400", errorResponse)

	statusAACResult := openapi3.NewResponse().
		WithDescription("Admin acp status result").
		WithJSONSchemaRef(statusAACResultSchema)
	statusAAC := openapi3.NewOperation()
	statusAAC.OperationID = "Check status of aac"
	statusAAC.Description = "Check status of admin acp system"
	statusAAC.Tags = []string{"acp_aac_status"}
	statusAAC.Responses = openapi3.NewResponses()
	statusAAC.AddResponse(200, statusAACResult)
	statusAAC.Responses.Set("400", errorResponse)

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
		"/acp/aac/relationship",
		http.MethodPost,
		addActorRelationshipAAC,
		h.AddAACActorRelationship,
	)
	router.AddRoute(
		"/acp/aac/relationship",
		http.MethodDelete,
		deleteActorRelationshipAAC,
		h.DeleteAACActorRelationship,
	)
	router.AddRoute(
		"/acp/aac/re-enable",
		http.MethodPost,
		reEnableAAC,
		h.ReEnableAAC,
	)
	router.AddRoute(
		"/acp/aac/disable",
		http.MethodPost,
		disableAAC,
		h.DisableAAC,
	)
	router.AddRoute(
		"/acp/aac/status",
		http.MethodGet,
		statusAAC,
		h.GetAACStatus,
	)
}

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
	"encoding/json"
	"io"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
)

type acpHandler struct{}

func (s *acpHandler) AddPolicy(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

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
	db := mustGetContextClientDB(req)

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
	db := mustGetContextClientDB(req)

	// Extract the "parameters" query parameter
	queryParams := req.URL.Query().Get("parameters")
	if queryParams == "" {
		responseJSON(rw, http.StatusBadRequest, errorResponse{NewErrMissingQueryParameter("parameters")})
		return
	}

	// Parse JSON from the query parameter
	var message deleteDocActorRelationshipRequest
	err := json.Unmarshal([]byte(queryParams), &message)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrInvalidQueryParamJSON})
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

	acpDeleteDocActorRelationshipResult := openapi3.NewResponse().
		WithDescription("Delete acp relationship result").
		WithJSONSchemaRef(acpRelationshipDeleteResultSchema)
	acpDeleteDocActorRelationship := openapi3.NewOperation()
	acpDeleteDocActorRelationship.OperationID = "delete relationship"
	acpDeleteDocActorRelationship.Description = "Delete an actor relationship using acp system"
	acpDeleteDocActorRelationship.Tags = []string{"acp_relationship"}
	acpDeleteDocActorRelationshipParam := &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        "parameters",
			In:          "query",
			Description: "Parameters for the delete request",
			Required:    true,
			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type:   openapi3.NewStringSchema().Type,
					Format: "",
					Example: `{
						"CollectionName": "Users",
						"DocID": "bae-9d443d0c-52f6-568b-8f74-e8ff0825697b",
						"Relation": "owner",
						"TargetActor": "did:key:z7r8oqkfiiVe4bHLYBjHZTJqGiUqCuMo6q7qiNGNYogBb8CZhDZ6RmFocZYYrsxCLew1E9bdWJ5tC7bVCGosfQDrSy7nf"
					}`,
				},
			},
		},
	}
	acpDeleteDocActorRelationship.AddParameter(acpDeleteDocActorRelationshipParam.Value)
	acpDeleteDocActorRelationship.Responses = openapi3.NewResponses()
	acpDeleteDocActorRelationship.AddResponse(200, acpDeleteDocActorRelationshipResult)
	acpDeleteDocActorRelationship.Responses.Set("400", errorResponse)

	router.AddRoute("/acp/policy", http.MethodPost, acpAddPolicy, h.AddPolicy)
	router.AddRoute("/acp/relationship", http.MethodPost, acpAddDocActorRelationship, h.AddDocActorRelationship)
	router.AddRoute("/acp/relationship", http.MethodDelete, acpDeleteDocActorRelationship, h.DeleteDocActorRelationship)
}

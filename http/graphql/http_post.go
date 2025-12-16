// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// primarily based on https://github.com/99designs/gqlgen/blob/master/graphql/handler/transport/http_post.go
package graphql

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
)

// POST implements the POST HTTP transport
// defined in https://github.com/APIs-guru/graphql-over-http#post
type POST struct{}

var _ Transport = POST{}

func (h POST) Supports(r *http.Request) bool {
	if r.Header.Get("Upgrade") != "" {
		return false
	}

	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return false
	}

	return r.Method == http.MethodPost && mediaType == "application/json"
}

// func getRequestBody(r *http.Request) (string, error) {
// 	if r == nil || r.Body == nil {
// 		return "", nil
// 	}
// 	body, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		return "", fmt.Errorf("unable to get Request Body %w", err)
// 	}
// 	return string(body), nil
// }

func (h POST) Do(w http.ResponseWriter, r *http.Request, executer Executor) {
	contentType := determineResponseContentType(r)
	responseHeaders := map[string][]string{
		"Content-Type": {contentType},
	}
	writeHeaders(w, responseHeaders)

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		gqlErr := errors.Wrap("could not read request body: %w", err)
		responseJSON(w, http.StatusBadRequest, errorResponse{gqlErr})
		return
	}

	var request GraphQLRequest
	bodyReader := bytes.NewReader(bodyBytes)
	err = json.NewDecoder(bodyReader).Decode(&request)
	if err != nil {
		gqlErr := errors.Wrap(
			"json request body could not be decoded: %w",
			err,
		)
		responseJSON(w, http.StatusBadRequest, errorResponse{gqlErr})
		return
	}

	var options []client.RequestOption
	if request.OperationName != "" {
		options = append(options, client.WithOperationName(request.OperationName))
	}
	if len(request.Variables) > 0 {
		options = append(options, client.WithVariables(request.Variables))
	}

	result, _, err := executer(r.Context(), request.Query, options...)
	if err != nil {
		gqlErr := errors.Wrap("could not execute request: %w", err)
		responseJSON(w, http.StatusBadRequest, errorResponse{gqlErr})
		return
	}

	responseJSON(w, http.StatusOK, result)
}

func (h POST) Operation() *openapi3.Operation {
	graphQLRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/graphql_request",
	}
	graphQLResponseSchema := openapi3.NewObjectSchema().
		WithProperties(map[string]*openapi3.Schema{
			"errors": openapi3.NewArraySchema().WithItems(
				openapi3.NewObjectSchema().WithProperties(map[string]*openapi3.Schema{
					"message": openapi3.NewStringSchema(),
				}),
			),
			"data": openapi3.NewObjectSchema().WithAnyAdditionalProperties(),
		})

	graphQLRequest := openapi3.NewRequestBody().
		WithContent(openapi3.NewContentWithJSONSchemaRef(graphQLRequestSchema))

	graphQLResponse := openapi3.NewResponse().
		WithDescription("GraphQL response").
		WithContent(openapi3.NewContentWithJSONSchema(graphQLResponseSchema))

	graphQLPost := openapi3.NewOperation()
	graphQLPost.Description = "GraphQL POST endpoint"
	graphQLPost.OperationID = "graphql_post"
	graphQLPost.Tags = []string{"graphql"}
	graphQLPost.RequestBody = &openapi3.RequestBodyRef{
		Value: graphQLRequest,
	}
	graphQLPost.AddResponse(200, graphQLResponse)
	graphQLPost.AddResponse(400, graphQLResponse)
	return graphQLPost
}

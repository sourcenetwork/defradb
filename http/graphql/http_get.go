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
	"encoding/json"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/sourcenetwork/defradb/client"
)

// GET implements the GET HTTP transport
// defined in https://github.com/APIs-guru/graphql-over-http#post
type GET struct{}

var _ Transport = GET{}

func (h GET) Supports(r *http.Request) bool {
	if r.Header.Get("Upgrade") != "" {
		return false
	}

	return r.Method == http.MethodGet
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

func (h GET) Do(w http.ResponseWriter, r *http.Request, executer Executor) {
	contentType := determineResponseContentType(r)
	responseHeaders := map[string][]string{
		"Content-Type": {contentType},
	}
	writeHeaders(w, responseHeaders)

	var request Request
	request.Query = r.URL.Query().Get("query")
	request.OperationName = r.URL.Query().Get("operationName")
	variablesFromQuery := r.URL.Query().Get("variables")
	if variablesFromQuery != "" {
		var variables map[string]any
		if err := json.Unmarshal([]byte(variablesFromQuery), &variables); err != nil {
			responseJSON(w, http.StatusNotAcceptable, errorResponse{ErrBadFormattedVariables})
			return
		}
		request.Variables = variables
	}

	var options []client.RequestOption
	if request.OperationName != "" {
		options = append(options, client.WithOperationName(request.OperationName))
	}
	if len(request.Variables) > 0 {
		options = append(options, client.WithVariables(request.Variables))
	}

	result := executer(r.Context(), request.Query, options...)
	if result.Subscription != nil {
		responseJSON(w, http.StatusNotAcceptable, errorResponse{ErrInvalidSubscriptionTransport})
	}
	responseJSON(w, http.StatusOK, result.GQL)
}

func (h GET) Methods() []string {
	return []string{http.MethodGet}
}

func (h GET) OpenAPI3Operation() *openapi3.Operation {
	graphQLResponseSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/graphql_response",
	}

	graphQLResponse := openapi3.NewResponse().
		WithDescription("GraphQL response").
		WithContent(openapi3.NewContentWithJSONSchemaRef(graphQLResponseSchema))

	graphQLQueryParam := openapi3.NewQueryParameter("query").
		WithSchema(openapi3.NewStringSchema())

	graphQLGet := openapi3.NewOperation()
	graphQLGet.Description = "GraphQL GET endpoint"
	graphQLGet.OperationID = "graphql_get"
	graphQLGet.Tags = []string{"graphql"}
	graphQLGet.AddParameter(graphQLQueryParam)
	graphQLGet.AddResponse(200, graphQLResponse)
	graphQLGet.AddResponse(400, graphQLResponse)
	return graphQLGet
}

/* GET

graphQLQueryParam := openapi3.NewQueryParameter("query").
		WithSchema(openapi3.NewStringSchema())

	graphQLGet := openapi3.NewOperation()
	graphQLGet.Description = "GraphQL GET endpoint"
	graphQLGet.OperationID = "graphql_get"
	graphQLGet.Tags = []string{"graphql"}
	graphQLGet.AddParameter(graphQLQueryParam)
	graphQLGet.AddResponse(200, graphQLResponse)
	graphQLGet.Responses.Set("400", errorResponse)

*/

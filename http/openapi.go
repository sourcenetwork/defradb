// Copyright 2023 Democratized Data Foundation
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
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"

	"github.com/sourcenetwork/defradb/client"
)

// openApiSchemas is a mapping of types to auto generate schemas for.
var openApiSchemas = map[string]any{
	"error":             &errorResponse{},
	"create_tx":         &CreateTxResponse{},
	"collection_update": &CollectionUpdateRequest{},
	"collection_delete": &CollectionDeleteRequest{},
	"peer_info":         &PeerInfoResponse{},
	"graphql_request":   &GraphQLRequest{},
	"graphql_response":  &GraphQLResponse{},
	"backup_config":     &client.BackupConfig{},
	"collection":        &client.CollectionDescription{},
	"index":             &client.IndexDescription{},
	"delete_result":     &client.DeleteResult{},
	"update_result":     &client.UpdateResult{},
	"lens_config":       &client.LensConfig{},
	"replicator":        &client.Replicator{},
	"ccip_request":      &CCIPRequest{},
	"ccip_response":     &CCIPResponse{},
}

func NewOpenAPISpec() (*openapi3.T, error) {
	schemas := make(openapi3.Schemas)
	responses := make(openapi3.Responses)
	parameters := make(openapi3.ParametersMap)

	generator := openapi3gen.NewGenerator(openapi3gen.UseAllExportedFields())
	for key, val := range openApiSchemas {
		ref, err := generator.NewSchemaRefForValue(val, schemas)
		if err != nil {
			return nil, err
		}
		schemas[key] = ref
	}

	errorSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/error",
	}

	errorResponse := openapi3.NewResponse().
		WithDescription("error").
		WithContent(openapi3.NewContentWithJSONSchemaRef(errorSchema))

	successResponse := openapi3.NewResponse().
		WithDescription("ok")

	txnHeaderParam := openapi3.NewHeaderParameter("x-defradb-tx").
		WithDescription("Transaction id").
		WithSchema(openapi3.NewInt64Schema())

	// add common schemas, responses, and params so we can reference them
	schemas["document"] = &openapi3.SchemaRef{
		Value: openapi3.NewObjectSchema().WithAnyAdditionalProperties(),
	}
	responses["success"] = &openapi3.ResponseRef{
		Value: successResponse,
	}
	responses["error"] = &openapi3.ResponseRef{
		Value: errorResponse,
	}
	parameters["txn"] = &openapi3.ParameterRef{
		Value: txnHeaderParam,
	}

	return &openapi3.T{
		OpenAPI: "3.0.3",
		Info: &openapi3.Info{
			Title:   "DefraDB API",
			Version: "0",
		},
		Paths: make(openapi3.Paths),
		Servers: openapi3.Servers{
			&openapi3.Server{
				Description: "Local DefraDB instance",
				URL:         "http://localhost:9181/api/v0",
			},
		},
		ExternalDocs: &openapi3.ExternalDocs{
			Description: "Read more about DefraDB",
			URL:         "https://docs.source.network",
		},
		Components: &openapi3.Components{
			Schemas:    schemas,
			Responses:  responses,
			Parameters: parameters,
		},
	}, nil
}

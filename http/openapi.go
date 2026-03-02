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

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
)

// openApiSchemas is a mapping of types to auto generate schemas for.
var openApiSchemas = map[string]any{
	"error":                                    &errorResponse{},
	"create_tx":                                &CreateTxResponse{},
	"update_collection":                        &UpdateCollectionRequest{},
	"delete_collection":                        &DeleteCollectionRequest{},
	"get_peer_info":                            &client.PeerInfo{},
	"request_graphql":                          &GraphQLRequest{},
	"backup_config":                            &client.BackupConfig{},
	"collection":                               &client.CollectionVersion{},
	"index":                                    &client.IndexDescription{},
	"new_index":                                &client.NewIndexRequest{},
	"encrypted_index":                          &client.EncryptedIndexDescription{},
	"add_encrypted_index":                      &client.EncryptedIndexDescription{},
	"delete_result":                            &client.DeleteResult{},
	"update_result":                            &client.UpdateResult{},
	"lens_config":                              &client.LensConfig{},
	"replicator":                               &client.Replicator{},
	"add_replicator_params":                    &AddReplicatorParams{},
	"delete_replicator_params":                 &DeleteReplicatorParams{},
	"ccip_request":                             &CCIPRequest{},
	"ccip_response":                            &CCIPResponse{},
	"patch_collection_request":                 &patchCollectionRequest{},
	"add_view_request":                         &addViewRequest{},
	"add_acp_policy_result":                    &client.AddPolicyResult{},
	"add_acp_relationship_result":              &client.AddActorRelationshipResult{},
	"delete_acp_relationship_result":           &client.DeleteActorRelationshipResult{},
	"get_acp_node_status_result":               &client.NACStatusResult{},
	"add_acp_node_relationship_request":        &addNACActorRelationshipRequest{},
	"delete_acp_node_relationship_request":     &deleteNACActorRelationshipRequest{},
	"add_acp_document_relationship_request":    &addDACActorRelationshipRequest{},
	"delete_acp_document_relationship_request": &deleteDACActorRelationshipRequest{},
	"identity":                                 &identity.PublicRawIdentity{},
	"set_migration":                            &SetMigrationResponse{},
	"add_lens_request":                         &AddLensRequest{},
	"add_lens_response":                        &AddLensResponse{},
}

func NewOpenAPISpec() (*openapi3.T, error) {
	schemas := make(openapi3.Schemas)
	responses := make(openapi3.ResponseBodies)
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

	// add authentication schemes
	securitySchemes := openapi3.SecuritySchemes{
		"bearerToken": &openapi3.SecuritySchemeRef{
			Value: openapi3.NewJWTSecurityScheme(),
		},
	}

	return &openapi3.T{
		OpenAPI: "3.0.3",
		Info: &openapi3.Info{
			Title:   "DefraDB API",
			Version: "0",
		},
		Paths: openapi3.NewPaths(),
		Servers: openapi3.Servers{
			&openapi3.Server{
				Description: "Local DefraDB instance",
				URL:         "/api/v0",
			},
		},
		ExternalDocs: &openapi3.ExternalDocs{
			Description: "Learn more about DefraDB",
			URL:         "https://docs.source.network",
		},
		Components: &openapi3.Components{
			Schemas:         schemas,
			Responses:       responses,
			Parameters:      parameters,
			SecuritySchemes: securitySchemes,
		},
		Tags: openapi3.Tags{
			&openapi3.Tag{
				Name:        "collection",
				Description: "Add, describe, patch, and manage collections",
			},
			&openapi3.Tag{
				Name:        "document",
				Description: "Add, get, update, or delete documents",
			},
			&openapi3.Tag{
				Name:        "view",
				Description: "Add views",
			},
			&openapi3.Tag{
				Name:        "index",
				Description: "Add, update, or remove indexes",
			},
			&openapi3.Tag{
				Name:        "lens",
				Description: "Migrate documents to and from collection versions",
			},
			&openapi3.Tag{
				Name:        "p2p",
				Description: "Peer-to-peer network operations",
			},
			&openapi3.Tag{
				Name:        "acp",
				Description: "Access control system(s)",
			},
			&openapi3.Tag{
				Name:        "transaction",
				Description: "Database transaction operations",
			},
			&openapi3.Tag{
				Name:        "backup",
				Description: "Database backup operations",
			},
			&openapi3.Tag{
				Name:        "graphql",
				Description: "GraphQL query endpoints",
			},
			&openapi3.Tag{
				Name: "ccip",
				ExternalDocs: &openapi3.ExternalDocs{
					Description: "EIP-3668",
					URL:         "https://eips.ethereum.org/EIPS/eip-3668",
				},
			},
		},
	}, nil
}

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
	"context"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"

	"github.com/sourcenetwork/defradb/client"
)

// OpenApiSpec contains the OpenAPI specification for DefraDB.
var OpenApiSpec = &openapi3.T{
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
}

func init() {
	generator := openapi3gen.NewGenerator(openapi3gen.UseAllExportedFields())

	errorSchema, err := generator.NewSchemaRefForValue(&errorResponse{}, nil)
	if err != nil {
		panic(err)
	}
	createTxSchema, err := generator.NewSchemaRefForValue(&CreateTxResponse{}, nil)
	if err != nil {
		panic(err)
	}
	backupConfigSchema, err := generator.NewSchemaRefForValue(&client.BackupConfig{}, nil)
	if err != nil {
		panic(err)
	}
	collectionSchema, err := generator.NewSchemaRefForValue(&client.CollectionDescription{}, nil)
	if err != nil {
		panic(err)
	}
	collectionUpdateSchema, err := generator.NewSchemaRefForValue(&CollectionUpdateRequest{}, nil)
	if err != nil {
		panic(err)
	}
	collectionDeleteSchema, err := generator.NewSchemaRefForValue(&CollectionDeleteRequest{}, nil)
	if err != nil {
		panic(err)
	}
	deleteResultSchema, err := generator.NewSchemaRefForValue(&client.DeleteResult{}, nil)
	if err != nil {
		panic(err)
	}
	updateResultSchema, err := generator.NewSchemaRefForValue(&client.UpdateResult{}, nil)
	if err != nil {
		panic(err)
	}
	indexDescriptionSchema, err := generator.NewSchemaRefForValue(&client.IndexDescription{}, nil)
	if err != nil {
		panic(err)
	}
	lensConfigSchema, err := generator.NewSchemaRefForValue(&client.LensConfig{}, nil)
	if err != nil {
		panic(err)
	}
	peerInfoSchema, err := generator.NewSchemaRefForValue(&PeerInfoResponse{}, nil)
	if err != nil {
		panic(err)
	}
	replicatorSchema, err := generator.NewSchemaRefForValue(&client.Replicator{}, nil)
	if err != nil {
		panic(err)
	}
	graphQLRequestSchema, err := generator.NewSchemaRefForValue(&GraphQLRequest{}, nil)
	if err != nil {
		panic(err)
	}
	graphQLResponseSchema, err := generator.NewSchemaRefForValue(&GraphQLResponse{}, nil)
	if err != nil {
		panic(err)
	}

	successResponse := openapi3.NewResponse().
		WithDescription("ok")

	errorResponse := openapi3.NewResponse().
		WithDescription("error").
		WithContent(openapi3.NewContentWithJSONSchemaRef(errorSchema))

	txnHeaderParam := openapi3.NewHeaderParameter("x-defradb-tx").
		WithDescription("Transaction id").
		WithSchema(openapi3.NewInt64Schema())

	txnReadOnlyQueryParam := openapi3.NewQueryParameter("read_only").
		WithDescription("Read only transaction").
		WithSchema(openapi3.NewBoolSchema().WithDefault(false))

	txnCreateResponse := openapi3.NewResponse().
		WithDescription("Transaction info").
		WithJSONSchemaRef(createTxSchema)

	txnCreate := openapi3.NewOperation()
	txnCreate.AddParameter(txnReadOnlyQueryParam)
	txnCreate.AddResponse(200, txnCreateResponse)
	txnCreate.AddResponse(400, errorResponse)

	txnConcurrent := openapi3.NewOperation()
	txnConcurrent.AddParameter(txnReadOnlyQueryParam)
	txnConcurrent.AddResponse(200, txnCreateResponse)
	txnConcurrent.AddResponse(400, errorResponse)

	txnIdPathParam := openapi3.NewPathParameter("id").
		WithRequired(true).
		WithSchema(openapi3.NewInt64Schema())

	txnCommit := openapi3.NewOperation()
	txnCommit.AddParameter(txnIdPathParam)
	txnCommit.AddResponse(200, successResponse)
	txnCommit.AddResponse(400, errorResponse)

	txnDiscard := openapi3.NewOperation()
	txnDiscard.AddParameter(txnIdPathParam)
	txnDiscard.AddResponse(200, successResponse)
	txnDiscard.AddResponse(400, errorResponse)

	backupRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithJSONSchemaRef(backupConfigSchema)

	backupExport := openapi3.NewOperation()
	backupExport.AddParameter(txnHeaderParam)
	backupExport.AddResponse(200, successResponse)
	backupExport.AddResponse(400, errorResponse)
	backupExport.RequestBody = &openapi3.RequestBodyRef{
		Value: backupRequest,
	}

	backupImport := openapi3.NewOperation()
	backupImport.AddParameter(txnHeaderParam)
	backupImport.AddResponse(200, successResponse)
	backupImport.AddResponse(400, errorResponse)
	backupImport.RequestBody = &openapi3.RequestBodyRef{
		Value: backupRequest,
	}

	collectionNameQueryParam := openapi3.NewQueryParameter("name").
		WithDescription("Collection name").
		WithSchema(openapi3.NewStringSchema())
	collectionSchemaIdQueryParam := openapi3.NewQueryParameter("schema_id").
		WithDescription("Collection schema id").
		WithSchema(openapi3.NewStringSchema())
	collectionVersionIdQueryParam := openapi3.NewQueryParameter("version_id").
		WithDescription("Collection schema version id").
		WithSchema(openapi3.NewStringSchema())

	collectionsSchema := openapi3.NewOneOfSchema(
		collectionSchema.Value,
		openapi3.NewArraySchema().WithItems(collectionSchema.Value),
	)
	collectionsResponse := openapi3.NewResponse().
		WithDescription("Collection(s) with matching name, schema id, or version id.").
		WithJSONSchema(collectionsSchema)

	collectionsList := openapi3.NewOperation()
	collectionsList.Description = "Get collection(s) by name, schema id, or version id."
	collectionsList.AddParameter(txnHeaderParam)
	collectionsList.AddParameter(collectionNameQueryParam)
	collectionsList.AddParameter(collectionSchemaIdQueryParam)
	collectionsList.AddParameter(collectionVersionIdQueryParam)
	collectionsList.AddResponse(200, collectionsResponse)
	collectionsList.AddResponse(400, errorResponse)

	collectionNamePathParam := openapi3.NewPathParameter("name").
		WithDescription("Collection name").
		WithRequired(true).
		WithSchema(openapi3.NewStringSchema())

	documentSchema := openapi3.NewObjectSchema().
		WithAnyAdditionalProperties()

	collectionCreateSchema := openapi3.NewOneOfSchema(
		documentSchema,
		openapi3.NewArraySchema().WithItems(documentSchema),
	)
	collectionCreateRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchema(collectionCreateSchema))

	collectionCreate := openapi3.NewOperation()
	collectionCreate.AddParameter(txnHeaderParam)
	collectionCreate.AddParameter(collectionNamePathParam)
	collectionCreate.RequestBody = &openapi3.RequestBodyRef{
		Value: collectionCreateRequest,
	}
	collectionCreate.AddResponse(200, successResponse)
	collectionCreate.AddResponse(400, errorResponse)

	collectionUpdateWithRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(collectionUpdateSchema))
	collectionUpdateWithResponse := openapi3.NewResponse().
		WithDescription("Update results").
		WithJSONSchemaRef(updateResultSchema)

	collectionUpdateWith := openapi3.NewOperation()
	collectionUpdateWith.Description = "Update document(s) in a collection"
	collectionUpdateWith.AddParameter(txnHeaderParam)
	collectionUpdateWith.AddParameter(collectionNamePathParam)
	collectionUpdateWith.RequestBody = &openapi3.RequestBodyRef{
		Value: collectionUpdateWithRequest,
	}
	collectionUpdateWith.AddResponse(200, collectionUpdateWithResponse)
	collectionUpdateWith.AddResponse(400, errorResponse)

	collectionDeleteWithRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(collectionDeleteSchema))
	collectionDeleteWithResponse := openapi3.NewResponse().
		WithDescription("Delete results").
		WithJSONSchemaRef(deleteResultSchema)

	collectionDeleteWith := openapi3.NewOperation()
	collectionDeleteWith.Description = "Delete document(s) from a collection"
	collectionDeleteWith.AddParameter(txnHeaderParam)
	collectionDeleteWith.AddParameter(collectionNamePathParam)
	collectionDeleteWith.RequestBody = &openapi3.RequestBodyRef{
		Value: collectionDeleteWithRequest,
	}
	collectionDeleteWith.AddResponse(200, collectionDeleteWithResponse)
	collectionDeleteWith.AddResponse(400, errorResponse)

	createIndexRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(indexDescriptionSchema))
	createIndexResponse := openapi3.NewResponse().
		WithDescription("Index description").
		WithJSONSchemaRef(indexDescriptionSchema)

	createIndex := openapi3.NewOperation()
	createIndex.AddParameter(txnHeaderParam)
	createIndex.AddParameter(collectionNamePathParam)
	createIndex.RequestBody = &openapi3.RequestBodyRef{
		Value: createIndexRequest,
	}
	createIndex.AddResponse(200, createIndexResponse)
	createIndex.AddResponse(400, errorResponse)

	getIndexesResponse := openapi3.NewResponse().
		WithDescription("List of indexes").
		WithJSONSchema(openapi3.NewArraySchema().WithItems(indexDescriptionSchema.Value))

	getIndexes := openapi3.NewOperation()
	getIndexes.AddParameter(txnHeaderParam)
	getIndexes.AddParameter(collectionNamePathParam)
	getIndexes.AddResponse(200, getIndexesResponse)
	getIndexes.AddResponse(400, errorResponse)

	indexPathParam := openapi3.NewPathParameter("index").
		WithRequired(true).
		WithSchema(openapi3.NewStringSchema())

	dropIndex := openapi3.NewOperation()
	dropIndex.AddParameter(txnHeaderParam)
	dropIndex.AddParameter(collectionNamePathParam)
	dropIndex.AddParameter(indexPathParam)
	dropIndex.AddResponse(200, successResponse)
	dropIndex.AddResponse(400, errorResponse)

	documentKeyPathParam := openapi3.NewPathParameter("key").
		WithRequired(true).
		WithSchema(openapi3.NewStringSchema())

	collectionGetResponse := openapi3.NewResponse().
		WithDescription("Document value").
		WithJSONSchema(documentSchema)

	collectionGet := openapi3.NewOperation()
	collectionGet.AddParameter(txnHeaderParam)
	collectionGet.AddParameter(collectionNamePathParam)
	collectionGet.AddParameter(documentKeyPathParam)
	collectionGet.AddResponse(200, collectionGetResponse)
	collectionGet.AddResponse(400, errorResponse)

	collectionUpdate := openapi3.NewOperation()
	collectionUpdate.AddParameter(txnHeaderParam)
	collectionUpdate.AddParameter(collectionNamePathParam)
	collectionUpdate.AddParameter(documentKeyPathParam)
	collectionUpdate.AddResponse(200, successResponse)
	collectionUpdate.AddResponse(400, errorResponse)

	collectionDelete := openapi3.NewOperation()
	collectionDelete.AddParameter(txnHeaderParam)
	collectionDelete.AddParameter(collectionNamePathParam)
	collectionDelete.AddParameter(documentKeyPathParam)
	collectionDelete.AddResponse(200, successResponse)
	collectionDelete.AddResponse(400, errorResponse)

	lensConfigResponse := openapi3.NewResponse().
		WithDescription("Lens configurations").
		WithJSONSchema(openapi3.NewArraySchema().WithItems(lensConfigSchema.Value))

	lensConfig := openapi3.NewOperation()
	lensConfig.AddParameter(txnHeaderParam)
	lensConfig.AddResponse(200, lensConfigResponse)
	lensConfig.AddResponse(400, errorResponse)

	setMigrationRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithJSONSchema(lensConfigSchema.Value)

	setMigration := openapi3.NewOperation()
	setMigration.AddParameter(txnHeaderParam)
	setMigration.RequestBody = &openapi3.RequestBodyRef{
		Value: setMigrationRequest,
	}
	setMigration.AddResponse(200, successResponse)
	setMigration.AddResponse(400, errorResponse)

	reloadLenses := openapi3.NewOperation()
	reloadLenses.AddParameter(txnHeaderParam)
	reloadLenses.AddResponse(200, successResponse)
	reloadLenses.AddResponse(400, errorResponse)

	versionPathParam := openapi3.NewPathParameter("version").
		WithRequired(true).
		WithSchema(openapi3.NewStringSchema())

	hasMigration := openapi3.NewOperation()
	hasMigration.AddParameter(txnHeaderParam)
	hasMigration.AddParameter(versionPathParam)
	hasMigration.AddResponse(200, successResponse)
	hasMigration.AddResponse(400, errorResponse)

	migrateSchema := openapi3.NewArraySchema().
		WithItems(documentSchema)
	migrateRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchema(migrateSchema))

	migrateUp := openapi3.NewOperation()
	migrateUp.AddParameter(txnHeaderParam)
	migrateUp.RequestBody = &openapi3.RequestBodyRef{
		Value: migrateRequest,
	}
	migrateUp.AddParameter(versionPathParam)
	migrateUp.AddResponse(200, successResponse)
	migrateUp.AddResponse(400, errorResponse)

	migrateDown := openapi3.NewOperation()
	migrateDown.AddParameter(txnHeaderParam)
	migrateDown.RequestBody = &openapi3.RequestBodyRef{
		Value: migrateRequest,
	}
	migrateDown.AddParameter(versionPathParam)
	migrateDown.AddResponse(200, successResponse)
	migrateDown.AddResponse(400, errorResponse)

	peerInfoResponse := openapi3.NewResponse().
		WithDescription("Peer network info").
		WithContent(openapi3.NewContentWithJSONSchemaRef(peerInfoSchema))

	peerInfo := openapi3.NewOperation()
	peerInfo.AddResponse(200, peerInfoResponse)
	peerInfo.AddResponse(400, errorResponse)

	getReplicatorsSchema := openapi3.NewArraySchema().
		WithItems(replicatorSchema.Value)
	getReplicatorsResponse := openapi3.NewResponse().
		WithDescription("Replicators").
		WithContent(openapi3.NewContentWithJSONSchema(getReplicatorsSchema))

	getReplicators := openapi3.NewOperation()
	getReplicators.AddResponse(200, getReplicatorsResponse)
	getReplicators.AddResponse(400, errorResponse)

	replicatorRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(replicatorSchema))

	setReplicator := openapi3.NewOperation()
	setReplicator.RequestBody = &openapi3.RequestBodyRef{
		Value: replicatorRequest,
	}
	setReplicator.AddResponse(200, successResponse)
	setReplicator.AddResponse(400, errorResponse)

	deleteReplicator := openapi3.NewOperation()
	deleteReplicator.RequestBody = &openapi3.RequestBodyRef{
		Value: replicatorRequest,
	}
	deleteReplicator.AddResponse(200, successResponse)
	deleteReplicator.AddResponse(400, errorResponse)

	peerCollectionsSchema := openapi3.NewArraySchema().
		WithItems(openapi3.NewStringSchema())

	peerCollectionRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchema(peerCollectionsSchema))

	getPeerCollectionsResponse := openapi3.NewResponse().
		WithDescription("Peer collections").
		WithContent(openapi3.NewContentWithJSONSchema(peerCollectionsSchema))

	getPeerCollections := openapi3.NewOperation()
	getPeerCollections.AddResponse(200, getPeerCollectionsResponse)
	getPeerCollections.AddResponse(400, errorResponse)

	addPeerCollections := openapi3.NewOperation()
	addPeerCollections.RequestBody = &openapi3.RequestBodyRef{
		Value: peerCollectionRequest,
	}
	addPeerCollections.AddResponse(200, successResponse)
	addPeerCollections.AddResponse(400, errorResponse)

	removePeerCollections := openapi3.NewOperation()
	removePeerCollections.RequestBody = &openapi3.RequestBodyRef{
		Value: peerCollectionRequest,
	}
	removePeerCollections.AddResponse(200, successResponse)
	removePeerCollections.AddResponse(400, errorResponse)

	graphQLQueryParam := openapi3.NewQueryParameter("query").
		WithSchema(openapi3.NewStringSchema())

	graphQLRequest := openapi3.NewRequestBody().
		WithContent(openapi3.NewContentWithJSONSchemaRef(graphQLRequestSchema))

	graphQLResponse := openapi3.NewResponse().
		WithDescription("GraphQL response").
		WithContent(openapi3.NewContentWithJSONSchemaRef(graphQLResponseSchema))

	graphQL := openapi3.NewOperation()
	graphQL.AddParameter(txnHeaderParam)
	graphQL.AddParameter(graphQLQueryParam)
	graphQL.RequestBody = &openapi3.RequestBodyRef{
		Value: graphQLRequest,
	}
	graphQL.AddResponse(200, graphQLResponse)
	graphQL.AddResponse(400, errorResponse)

	debugDump := openapi3.NewOperation()
	debugDump.AddResponse(200, successResponse)
	debugDump.AddResponse(400, errorResponse)

	OpenApiSpec.AddOperation("/txn", http.MethodPost, txnCreate)
	OpenApiSpec.AddOperation("/txn/concurrent", http.MethodPost, txnConcurrent)
	OpenApiSpec.AddOperation("/txn/{id}", http.MethodPost, txnCommit)
	OpenApiSpec.AddOperation("/txn/{id}", http.MethodDelete, txnDiscard)

	OpenApiSpec.AddOperation("/backup/export", http.MethodPost, backupExport)
	OpenApiSpec.AddOperation("/backup/import", http.MethodPost, backupImport)

	OpenApiSpec.AddOperation("/collections", http.MethodGet, collectionsList)
	OpenApiSpec.AddOperation("/collections/{name}", http.MethodPost, collectionCreate)
	OpenApiSpec.AddOperation("/collections/{name}", http.MethodPatch, collectionUpdateWith)
	OpenApiSpec.AddOperation("/collections/{name}", http.MethodDelete, collectionDeleteWith)
	OpenApiSpec.AddOperation("/collections/{name}/indexes", http.MethodPost, createIndex)
	OpenApiSpec.AddOperation("/collections/{name}/indexes", http.MethodGet, getIndexes)
	OpenApiSpec.AddOperation("/collections/{name}/indexes/{index}", http.MethodDelete, dropIndex)
	OpenApiSpec.AddOperation("/collections/{name}/{key}", http.MethodGet, collectionGet)
	OpenApiSpec.AddOperation("/collections/{name}/{key}", http.MethodPatch, collectionUpdate)
	OpenApiSpec.AddOperation("/collections/{name}/{key}", http.MethodDelete, collectionDelete)

	OpenApiSpec.AddOperation("/lens", http.MethodGet, lensConfig)
	OpenApiSpec.AddOperation("/lens", http.MethodPost, lensConfig)
	OpenApiSpec.AddOperation("/lens/reload", http.MethodPost, reloadLenses)
	OpenApiSpec.AddOperation("/lens/{version}", http.MethodGet, hasMigration)
	OpenApiSpec.AddOperation("/lens/{version}/up", http.MethodPost, migrateUp)
	OpenApiSpec.AddOperation("/lens/{version}/down", http.MethodPost, migrateDown)

	OpenApiSpec.AddOperation("/p2p/info", http.MethodGet, peerInfo)
	OpenApiSpec.AddOperation("/p2p/replicators", http.MethodGet, getReplicators)
	OpenApiSpec.AddOperation("/p2p/replicators", http.MethodPost, setReplicator)
	OpenApiSpec.AddOperation("/p2p/replicators", http.MethodDelete, deleteReplicator)
	OpenApiSpec.AddOperation("/p2p/collections", http.MethodGet, getPeerCollections)
	OpenApiSpec.AddOperation("/p2p/collections", http.MethodPost, addPeerCollections)
	OpenApiSpec.AddOperation("/p2p/collections", http.MethodDelete, removePeerCollections)

	OpenApiSpec.AddOperation("/graphql", http.MethodGet, graphQL)
	OpenApiSpec.AddOperation("/graphql", http.MethodPost, graphQL)

	OpenApiSpec.AddOperation("/debug/dump", http.MethodGet, debugDump)

	// ensure the specification is always valid
	if err := OpenApiSpec.Validate(context.Background()); err != nil {
		panic(err)
	}
}

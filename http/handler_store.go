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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/sourcenetwork/defradb/client"
)

type storeHandler struct{}

func (s *storeHandler) BasicImport(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	var config client.BackupConfig
	if err := requestJSON(req, &config); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	err := store.BasicImport(req.Context(), config.Filepath)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *storeHandler) BasicExport(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	var config client.BackupConfig
	if err := requestJSON(req, &config); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	err := store.BasicExport(req.Context(), &config)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *storeHandler) AddSchema(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	schema, err := io.ReadAll(req.Body)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	cols, err := store.AddSchema(req.Context(), string(schema))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, cols)
}

func (s *storeHandler) PatchSchema(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	var message patchSchemaRequest
	err := requestJSON(req, &message)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	err = store.PatchSchema(req.Context(), message.Patch, message.SetAsDefaultVersion)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *storeHandler) SetDefaultSchemaVersion(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	schemaVersionID, err := io.ReadAll(req.Body)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	err = store.SetDefaultSchemaVersion(req.Context(), string(schemaVersionID))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *storeHandler) GetCollection(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	switch {
	case req.URL.Query().Has("name"):
		col, err := store.GetCollectionByName(req.Context(), req.URL.Query().Get("name"))
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		responseJSON(rw, http.StatusOK, col.Definition())
	case req.URL.Query().Has("schema_root"):
		cols, err := store.GetCollectionsBySchemaRoot(req.Context(), req.URL.Query().Get("schema_root"))
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		colDesc := make([]client.CollectionDefinition, len(cols))
		for i, col := range cols {
			colDesc[i] = col.Definition()
		}
		responseJSON(rw, http.StatusOK, colDesc)
	case req.URL.Query().Has("version_id"):
		cols, err := store.GetCollectionsByVersionID(req.Context(), req.URL.Query().Get("version_id"))
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		colDesc := make([]client.CollectionDefinition, len(cols))
		for i, col := range cols {
			colDesc[i] = col.Definition()
		}
		responseJSON(rw, http.StatusOK, colDesc)
	default:
		cols, err := store.GetAllCollections(req.Context())
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		colDesc := make([]client.CollectionDefinition, len(cols))
		for i, col := range cols {
			colDesc[i] = col.Definition()
		}
		responseJSON(rw, http.StatusOK, colDesc)
	}
}

func (s *storeHandler) GetSchema(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	switch {
	case req.URL.Query().Has("name"):
		schema, err := store.GetSchemasByName(req.Context(), req.URL.Query().Get("name"))
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		responseJSON(rw, http.StatusOK, schema)
	case req.URL.Query().Has("root"):
		schema, err := store.GetSchemasByRoot(req.Context(), req.URL.Query().Get("root"))
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		responseJSON(rw, http.StatusOK, schema)
	case req.URL.Query().Has("version_id"):
		schema, err := store.GetSchemaByVersionID(req.Context(), req.URL.Query().Get("version_id"))
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		responseJSON(rw, http.StatusOK, schema)
	default:
		schema, err := store.GetAllSchemas(req.Context())
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		responseJSON(rw, http.StatusOK, schema)
	}
}

func (s *storeHandler) GetAllIndexes(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	indexes, err := store.GetAllIndexes(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, indexes)
}

func (s *storeHandler) PrintDump(rw http.ResponseWriter, req *http.Request) {
	db := req.Context().Value(dbContextKey).(client.DB)

	if err := db.PrintDump(req.Context()); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

type GraphQLRequest struct {
	Query string `json:"query"`
}

type GraphQLResponse struct {
	Data   any     `json:"data"`
	Errors []error `json:"errors,omitempty"`
}

func (res GraphQLResponse) MarshalJSON() ([]byte, error) {
	var errors []string
	for _, err := range res.Errors {
		errors = append(errors, err.Error())
	}
	return json.Marshal(map[string]any{"data": res.Data, "errors": errors})
}

func (res *GraphQLResponse) UnmarshalJSON(data []byte) error {
	// decode numbers to json.Number
	dec := json.NewDecoder(bytes.NewBuffer(data))
	dec.UseNumber()

	var out map[string]any
	if err := dec.Decode(&out); err != nil {
		return err
	}

	// fix errors type to match tests
	switch t := out["errors"].(type) {
	case []any:
		for _, v := range t {
			res.Errors = append(res.Errors, parseError(v))
		}
	default:
		res.Errors = nil
	}

	// fix data type to match tests
	switch t := out["data"].(type) {
	case []any:
		var fixed []map[string]any
		for _, v := range t {
			fixed = append(fixed, v.(map[string]any))
		}
		res.Data = fixed
	case map[string]any:
		res.Data = t
	default:
		res.Data = []map[string]any{}
	}

	return nil
}

func (s *storeHandler) ExecRequest(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	var request GraphQLRequest
	switch {
	case req.URL.Query().Get("query") != "":
		request.Query = req.URL.Query().Get("query")
	case req.Body != nil:
		if err := requestJSON(req, &request); err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
	default:
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrMissingRequest})
		return
	}
	result := store.ExecRequest(req.Context(), request.Query)

	if result.Pub == nil {
		responseJSON(rw, http.StatusOK, GraphQLResponse{result.GQL.Data, result.GQL.Errors})
		return
	}
	flusher, ok := rw.(http.Flusher)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrStreamingNotSupported})
		return
	}

	rw.Header().Add("Content-Type", "text/event-stream")
	rw.Header().Add("Cache-Control", "no-cache")
	rw.Header().Add("Connection", "keep-alive")

	rw.WriteHeader(http.StatusOK)
	flusher.Flush()

	for {
		select {
		case <-req.Context().Done():
			return
		case item, open := <-result.Pub.Stream():
			if !open {
				return
			}
			data, err := json.Marshal(item)
			if err != nil {
				return
			}
			fmt.Fprintf(rw, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

func (h *storeHandler) bindRoutes(router *Router) {
	successResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/success",
	}
	errorResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/error",
	}
	collectionSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/collection",
	}
	schemaSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/schema",
	}
	graphQLRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/graphql_request",
	}
	graphQLResponseSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/graphql_response",
	}
	backupConfigSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/backup_config",
	}
	patchSchemaRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/patch_schema_request",
	}

	collectionArraySchema := openapi3.NewArraySchema()
	collectionArraySchema.Items = collectionSchema

	addSchemaResponse := openapi3.NewResponse().
		WithDescription("Collection(s)").
		WithJSONSchema(collectionArraySchema)

	addSchemaRequest := openapi3.NewRequestBody().
		WithContent(openapi3.NewContentWithSchema(openapi3.NewStringSchema(), []string{"text/plain"}))

	addSchema := openapi3.NewOperation()
	addSchema.OperationID = "add_schema"
	addSchema.Description = "Add a new schema definition"
	addSchema.Tags = []string{"schema"}
	addSchema.RequestBody = &openapi3.RequestBodyRef{
		Value: addSchemaRequest,
	}
	addSchema.AddResponse(200, addSchemaResponse)
	addSchema.Responses["400"] = errorResponse

	patchSchemaRequest := openapi3.NewRequestBody().
		WithJSONSchemaRef(patchSchemaRequestSchema)

	patchSchema := openapi3.NewOperation()
	patchSchema.OperationID = "patch_schema"
	patchSchema.Description = "Update a schema definition"
	patchSchema.Tags = []string{"schema"}
	patchSchema.RequestBody = &openapi3.RequestBodyRef{
		Value: patchSchemaRequest,
	}
	patchSchema.Responses = make(openapi3.Responses)
	patchSchema.Responses["200"] = successResponse
	patchSchema.Responses["400"] = errorResponse

	setDefaultSchemaVersionRequest := openapi3.NewRequestBody().
		WithContent(openapi3.NewContentWithSchema(openapi3.NewStringSchema(), []string{"text/plain"}))

	setDefaultSchemaVersion := openapi3.NewOperation()
	setDefaultSchemaVersion.OperationID = "set_default_schema_version"
	setDefaultSchemaVersion.Description = "Set the default schema version for a collection"
	setDefaultSchemaVersion.Tags = []string{"schema"}
	setDefaultSchemaVersion.RequestBody = &openapi3.RequestBodyRef{
		Value: setDefaultSchemaVersionRequest,
	}
	setDefaultSchemaVersion.Responses = make(openapi3.Responses)
	setDefaultSchemaVersion.Responses["200"] = successResponse
	setDefaultSchemaVersion.Responses["400"] = errorResponse

	backupRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithJSONSchemaRef(backupConfigSchema)

	backupExport := openapi3.NewOperation()
	backupExport.OperationID = "backup_export"
	backupExport.Description = "Export a database backup to file"
	backupExport.Tags = []string{"backup"}
	backupExport.Responses = make(openapi3.Responses)
	backupExport.Responses["200"] = successResponse
	backupExport.Responses["400"] = errorResponse
	backupExport.RequestBody = &openapi3.RequestBodyRef{
		Value: backupRequest,
	}

	backupImport := openapi3.NewOperation()
	backupImport.OperationID = "backup_import"
	backupImport.Description = "Import a database backup from file"
	backupImport.Tags = []string{"backup"}
	backupImport.Responses = make(openapi3.Responses)
	backupImport.Responses["200"] = successResponse
	backupImport.Responses["400"] = errorResponse
	backupImport.RequestBody = &openapi3.RequestBodyRef{
		Value: backupRequest,
	}

	collectionNameQueryParam := openapi3.NewQueryParameter("name").
		WithDescription("Collection name").
		WithSchema(openapi3.NewStringSchema())
	collectionSchemaRootQueryParam := openapi3.NewQueryParameter("schema_root").
		WithDescription("Collection schema root").
		WithSchema(openapi3.NewStringSchema())
	collectionVersionIdQueryParam := openapi3.NewQueryParameter("version_id").
		WithDescription("Collection schema version id").
		WithSchema(openapi3.NewStringSchema())

	collectionsSchema := openapi3.NewArraySchema()
	collectionsSchema.Items = collectionSchema

	collectionResponseSchema := openapi3.NewOneOfSchema()
	collectionResponseSchema.OneOf = openapi3.SchemaRefs{
		collectionSchema,
		openapi3.NewSchemaRef("", collectionsSchema),
	}

	collectionsResponse := openapi3.NewResponse().
		WithDescription("Collection(s) with matching name, schema id, or version id.").
		WithJSONSchema(collectionResponseSchema)

	collectionDescribe := openapi3.NewOperation()
	collectionDescribe.OperationID = "collection_describe"
	collectionDescribe.Description = "Introspect collection(s) by name, schema id, or version id."
	collectionDescribe.Tags = []string{"collection"}
	collectionDescribe.AddParameter(collectionNameQueryParam)
	collectionDescribe.AddParameter(collectionSchemaRootQueryParam)
	collectionDescribe.AddParameter(collectionVersionIdQueryParam)
	collectionDescribe.AddResponse(200, collectionsResponse)
	collectionDescribe.Responses["400"] = errorResponse

	schemaNameQueryParam := openapi3.NewQueryParameter("name").
		WithDescription("Schema name").
		WithSchema(openapi3.NewStringSchema())
	schemaSchemaRootQueryParam := openapi3.NewQueryParameter("root").
		WithDescription("Schema root").
		WithSchema(openapi3.NewStringSchema())
	schemaVersionIDQueryParam := openapi3.NewQueryParameter("version_id").
		WithDescription("Schema version id").
		WithSchema(openapi3.NewStringSchema())

	schemasSchema := openapi3.NewArraySchema()
	schemasSchema.Items = schemaSchema

	schemaResponseSchema := openapi3.NewOneOfSchema()
	schemaResponseSchema.OneOf = openapi3.SchemaRefs{
		schemaSchema,
		openapi3.NewSchemaRef("", schemasSchema),
	}

	schemaResponse := openapi3.NewResponse().
		WithDescription("Schema(s) with matching name, schema id, or version id.").
		WithJSONSchema(schemaResponseSchema)

	schemaDescribe := openapi3.NewOperation()
	schemaDescribe.OperationID = "schema_describe"
	schemaDescribe.Description = "Introspect schema(s) by name, schema root, or version id."
	schemaDescribe.Tags = []string{"schema"}
	schemaDescribe.AddParameter(schemaNameQueryParam)
	schemaDescribe.AddParameter(schemaSchemaRootQueryParam)
	schemaDescribe.AddParameter(schemaVersionIDQueryParam)
	schemaDescribe.AddResponse(200, schemaResponse)
	schemaDescribe.Responses["400"] = errorResponse

	graphQLRequest := openapi3.NewRequestBody().
		WithContent(openapi3.NewContentWithJSONSchemaRef(graphQLRequestSchema))

	graphQLResponse := openapi3.NewResponse().
		WithDescription("GraphQL response").
		WithContent(openapi3.NewContentWithJSONSchemaRef(graphQLResponseSchema))

	graphQLPost := openapi3.NewOperation()
	graphQLPost.Description = "GraphQL POST endpoint"
	graphQLPost.OperationID = "graphql_post"
	graphQLPost.Tags = []string{"graphql"}
	graphQLPost.RequestBody = &openapi3.RequestBodyRef{
		Value: graphQLRequest,
	}
	graphQLPost.AddResponse(200, graphQLResponse)
	graphQLPost.Responses["400"] = errorResponse

	graphQLQueryParam := openapi3.NewQueryParameter("query").
		WithSchema(openapi3.NewStringSchema())

	graphQLGet := openapi3.NewOperation()
	graphQLGet.Description = "GraphQL GET endpoint"
	graphQLGet.OperationID = "graphql_get"
	graphQLGet.Tags = []string{"graphql"}
	graphQLGet.AddParameter(graphQLQueryParam)
	graphQLGet.AddResponse(200, graphQLResponse)
	graphQLGet.Responses["400"] = errorResponse

	debugDump := openapi3.NewOperation()
	debugDump.Description = "Dump database"
	debugDump.OperationID = "debug_dump"
	debugDump.Tags = []string{"debug"}
	debugDump.Responses = make(openapi3.Responses)
	debugDump.Responses["200"] = successResponse
	debugDump.Responses["400"] = errorResponse

	router.AddRoute("/backup/export", http.MethodPost, backupExport, h.BasicExport)
	router.AddRoute("/backup/import", http.MethodPost, backupImport, h.BasicImport)
	router.AddRoute("/collections", http.MethodGet, collectionDescribe, h.GetCollection)
	router.AddRoute("/graphql", http.MethodGet, graphQLGet, h.ExecRequest)
	router.AddRoute("/graphql", http.MethodPost, graphQLPost, h.ExecRequest)
	router.AddRoute("/debug/dump", http.MethodGet, debugDump, h.PrintDump)
	router.AddRoute("/schema", http.MethodPost, addSchema, h.AddSchema)
	router.AddRoute("/schema", http.MethodPatch, patchSchema, h.PatchSchema)
	router.AddRoute("/schema", http.MethodGet, schemaDescribe, h.GetSchema)
	router.AddRoute("/schema/default", http.MethodPost, setDefaultSchemaVersion, h.SetDefaultSchemaVersion)
}

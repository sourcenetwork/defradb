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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/identity"
)

const (
	sseAcceptHeader  = "text/event-stream"
	jsonAcceptHeader = "application/json"
)

type storeHandler struct{}

func (h *storeHandler) BasicImport(rw http.ResponseWriter, req *http.Request) {
	if !IsDevMode {
		responseJSON(rw, http.StatusBadRequest, errorResponse{client.NewErrOperationRequiresDeveloperMode("BasicImport")})
		return
	}

	db := mustGetContextClientDB(req)

	var config client.BackupConfig
	if err := requestJSON(req, &config); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	err := db.BasicImport(req.Context(), config.Filepath)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h *storeHandler) BasicExport(rw http.ResponseWriter, req *http.Request) {
	if !IsDevMode {
		responseJSON(rw, http.StatusBadRequest, errorResponse{client.NewErrOperationRequiresDeveloperMode("BasicExport")})
		return
	}

	db := mustGetContextClientDB(req)
	ctx := req.Context()

	var config client.BackupConfig
	if err := requestJSON(req, &config); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	opt := options.BasicExport().
		SetFormat(config.Format).
		SetPretty(config.Pretty).
		SetCollections(config.Collections)

	err := db.BasicExport(ctx, config.Filepath, opt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h *storeHandler) AddSchema(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()

	schema, err := io.ReadAll(req.Body)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	opt := options.WithIdentity(options.AddSchema(), identity.FromContext(ctx))
	cols, err := db.AddSchema(ctx, string(schema), opt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, cols)
}

func (h *storeHandler) PatchCollection(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()

	var message patchCollectionRequest
	err := requestJSON(req, &message)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	opt := options.WithIdentity(options.PatchCollection(), identity.FromContext(ctx))
	err = db.PatchCollection(ctx, message.Patch, message.Migration, opt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h *storeHandler) SetActiveCollectionVersion(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()

	collectionVersionID, err := io.ReadAll(req.Body)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	opt := options.WithIdentity(options.SetActiveCollectionVersion(), identity.FromContext(ctx))
	err = db.SetActiveCollectionVersion(ctx, string(collectionVersionID), opt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h *storeHandler) AddView(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()

	var message addViewRequest
	err := requestJSON(req, &message)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	opt := options.WithIdentity(options.AddView(), identity.FromContext(ctx))
	if message.TransformCID.HasValue() {
		opt.SetTransformCID(message.TransformCID.Value())
	}

	defs, err := db.AddView(ctx, message.Query, message.SDL, opt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	responseJSON(rw, http.StatusOK, defs)
}

type SetMigrationResponse struct {
	LensID string `json:"lensId"`
}

func (h *storeHandler) SetMigration(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	var cfg client.LensConfig
	if err := requestJSON(req, &cfg); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	opts := options.WithIdentity(options.SetMigration(), identity.FromContext(req.Context()))

	lensID, err := db.SetMigration(req.Context(), cfg, opts)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	responseJSON(rw, http.StatusOK, &SetMigrationResponse{LensID: lensID})
}

type AddLensRequest struct {
	Lens model.Lens `json:"lens"`
}

type AddLensResponse struct {
	LensID string `json:"lensId"`
}

func (h *storeHandler) AddLens(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	var addLensReq AddLensRequest
	if err := requestJSON(req, &addLensReq); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	opts := options.WithIdentity(options.AddLens(), identity.FromContext(req.Context()))

	lensID, err := db.AddLens(req.Context(), addLensReq.Lens, opts)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	responseJSON(rw, http.StatusOK, &AddLensResponse{LensID: lensID})
}

type ListLensesResponse struct {
	Lenses map[string]model.Lens `json:"lenses"`
}

func (h *storeHandler) ListLenses(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	opts := options.WithIdentity(options.ListLenses(), identity.FromContext(req.Context()))

	lenses, err := db.ListLenses(req.Context(), opts)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	responseJSON(rw, http.StatusOK, &ListLensesResponse{Lenses: lenses})
}

func (h *storeHandler) GetCollection(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()

	opt := options.WithIdentity(options.GetCollections(), identity.FromContext(ctx))
	if req.URL.Query().Has("name") {
		opt.SetCollectionName(req.URL.Query().Get("name"))
	}
	if req.URL.Query().Has("version_id") {
		opt.SetVersionID(req.URL.Query().Get("version_id"))
	}
	if req.URL.Query().Has("collection_id") {
		opt.SetCollectionID(req.URL.Query().Get("collection_id"))
	}
	if req.URL.Query().Has("get_inactive") {
		getInactiveStr := req.URL.Query().Get("get_inactive")
		var err error
		getInactive, err := strconv.ParseBool(getInactiveStr)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		opt.SetGetInactive(getInactive)
	}

	cols, err := db.GetCollections(ctx, opt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	colDesc := make([]client.CollectionVersion, len(cols))
	for i, col := range cols {
		colDesc[i] = col.Version()
	}
	responseJSON(rw, http.StatusOK, colDesc)
}

func (h *storeHandler) RefreshViews(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	opt := options.WithIdentity(options.RefreshViews(), identity.FromContext(req.Context()))
	if req.URL.Query().Has("name") {
		opt.SetCollectionName(req.URL.Query().Get("name"))
	}
	if req.URL.Query().Has("version_id") {
		opt.SetVersionID(req.URL.Query().Get("version_id"))
	}
	if req.URL.Query().Has("collection_id") {
		opt.SetCollectionID(req.URL.Query().Get("collection_id"))
	}
	if req.URL.Query().Has("get_inactive") {
		getInactiveStr := req.URL.Query().Get("get_inactive")
		getInactive, err := strconv.ParseBool(getInactiveStr)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		opt.SetGetInactive(getInactive)
	}

	err := db.RefreshViews(req.Context(), opt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h *storeHandler) ListIndexes(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	indexes, err := db.ListIndexes(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, indexes)
}

func (h *storeHandler) ListAllEncryptedIndexes(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	opts := options.WithIdentity(options.ListAllEncryptedIndexes(), identity.FromContext(req.Context()))
	indexes, err := db.ListAllEncryptedIndexes(req.Context(), opts)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, indexes)
}

func (h *storeHandler) PrintDump(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	if err := db.PrintDump(req.Context()); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

type GraphQLRequest struct {
	Query         string         `json:"query"`
	OperationName string         `json:"operationName"`
	Variables     map[string]any `json:"variables"`
}

func (h *storeHandler) ExecRequest(rw http.ResponseWriter, req *http.Request) {
	// handle different request transports
	// specifically, SSE
	if req.Header.Get("Accept") == sseAcceptHeader {
		execSSESubscription(rw, req)
		return
	}

	// if its not a subscription, then its just a regular
	// GraphQL over HTTP request
	execHTTPRequest(rw, req)
}

func execHTTPRequest(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()

	request, opts, err := extractGraphQLRequest(req)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	opts = options.WithIdentity(opts, identity.FromContext(ctx))
	result := db.ExecRequest(ctx, request.Query, opts)

	// if at this point the we get a subscription query, it isn't using
	// the correct accept headers, and we error
	if result.Subscription != nil {
		responseJSON(rw, http.StatusNotAcceptable, errorResponse{ErrInvalidSubscriptionTransport})
		return
	}

	responseJSON(rw, http.StatusOK, result.GQL)
}

func execSSESubscription(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()

	request, opts, err := extractGraphQLRequest(req)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	opts = options.WithIdentity(opts, identity.FromContext(ctx))

	// upgrade to SSE connection
	flusher, ok := rw.(http.Flusher)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrStreamingNotSupported})
		return
	}

	rw.Header().Add("Content-Type", sseAcceptHeader)
	rw.Header().Add("Cache-Control", "no-cache")
	rw.Header().Add("Connection", "keep-alive")
	rw.WriteHeader(http.StatusOK)
	flusher.Flush()

	result := db.ExecRequest(ctx, request.Query, opts)

	// if we get an error in the initial GQL request, we need to emit
	// it as a SSE event, then we can close the connection/subscription
	if len(result.GQL.Errors) > 0 {
		data, err := json.Marshal(result.GQL)
		if err != nil {
			return
		}

		err = emitSSENextEvent(rw, flusher, string(data))
		if err != nil {
			return
		}

		_ = emitSSECompleteEvent(rw, flusher)
		return
	}

	serverCtx, hasServerCtx := tryGetContexCtx(req)
	var serverDone <-chan struct{}
	if hasServerCtx {
		serverDone = serverCtx.Done()
	}
	for {
		select {
		case <-req.Context().Done():
			return
		case <-serverDone:
			// We need to check for closure of the server context
			// otherwise the server won't gracefully shutdown until all
			// connections are closed.
			_ = emitSSECompleteEvent(rw, flusher)
			return
		case item, open := <-result.Subscription:
			if !open {
				return
			}
			data, err := json.Marshal(item)
			if err != nil {
				return
			}

			_ = emitSSENextEvent(rw, flusher, string(data))
		}
	}
}

func emitSSENextEvent(rw http.ResponseWriter, flusher http.Flusher, data string) error {
	return emitSSEEvent(rw, flusher, "next", data)
}

func emitSSECompleteEvent(rw http.ResponseWriter, flusher http.Flusher) error {
	return emitSSEEvent(rw, flusher, "complete", "{}")
}

func emitSSEEvent(rw http.ResponseWriter, flusher http.Flusher, eventType string, data string) error {
	// For compatibility with SSE, the payload should have
	// a line defining the `event`.
	_, err := fmt.Fprintf(rw, "event: %s\n", eventType)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(rw, "data: %s\n\n", data)
	if err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

func extractGraphQLRequest(req *http.Request) (GraphQLRequest, *options.ExecRequestOptionsBuilder, error) {
	var request GraphQLRequest
	switch {
	case req.URL.Query().Get("query") != "":

		request.Query = req.URL.Query().Get("query")

		request.OperationName = req.URL.Query().Get("operationName")

		variablesFromQuery := req.URL.Query().Get("variables")
		if variablesFromQuery != "" {
			var variables map[string]any
			if err := json.Unmarshal([]byte(variablesFromQuery), &variables); err != nil {
				return GraphQLRequest{}, nil, err
			}
			request.Variables = variables
		}

	case req.Body != nil:
		if err := requestJSON(req, &request); err != nil {
			return GraphQLRequest{}, nil, err
		}
	default:
		return GraphQLRequest{}, nil, ErrMissingRequest
	}
	opt := options.ExecRequest()
	if request.OperationName != "" {
		opt.SetOperationName(request.OperationName)
	}
	if len(request.Variables) > 0 {
		opt.SetVariables(request.Variables)
	}

	return request, opt, nil
}

func (h *storeHandler) GetNodeIdentity(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	identity, err := db.GetNodeIdentity(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, identity)
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
	graphQLRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/graphql_request",
	}
	backupConfigSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/backup_config",
	}
	addViewSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/add_view_request",
	}
	lensConfigSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/lens_config",
	}
	patchCollectionRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/patch_collection_request",
	}
	identitySchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/identity",
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
	addSchema.Responses.Set("400", errorResponse)

	setActiveCollectionVersionRequest := openapi3.NewRequestBody().
		WithContent(openapi3.NewContentWithSchema(openapi3.NewStringSchema(), []string{"text/plain"}))

	setActiveCollectionVersion := openapi3.NewOperation()
	setActiveCollectionVersion.OperationID = "set_default_collection_version"
	setActiveCollectionVersion.Description = "Set the default version for a collection"
	setActiveCollectionVersion.Tags = []string{"collection"}
	setActiveCollectionVersion.RequestBody = &openapi3.RequestBodyRef{
		Value: setActiveCollectionVersionRequest,
	}
	setActiveCollectionVersion.Responses = openapi3.NewResponses()
	setActiveCollectionVersion.Responses.Set("200", successResponse)
	setActiveCollectionVersion.Responses.Set("400", errorResponse)

	backupRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithJSONSchemaRef(backupConfigSchema)

	backupExport := openapi3.NewOperation()
	backupExport.OperationID = "backup_export"
	backupExport.Description = "Export a database backup to file"
	backupExport.Tags = []string{"backup"}
	backupExport.Responses = openapi3.NewResponses()
	backupExport.Responses.Set("200", successResponse)
	backupExport.Responses.Set("400", errorResponse)
	backupExport.RequestBody = &openapi3.RequestBodyRef{
		Value: backupRequest,
	}

	backupImport := openapi3.NewOperation()
	backupImport.OperationID = "backup_import"
	backupImport.Description = "Import a database backup from file"
	backupImport.Tags = []string{"backup"}
	backupImport.Responses = openapi3.NewResponses()
	backupImport.Responses.Set("200", successResponse)
	backupImport.Responses.Set("400", errorResponse)
	backupImport.RequestBody = &openapi3.RequestBodyRef{
		Value: backupRequest,
	}

	collectionNameQueryParam := openapi3.NewQueryParameter("name").
		WithDescription("Collection name").
		WithSchema(openapi3.NewStringSchema())
	collectionSchemaRootQueryParam := openapi3.NewQueryParameter("collection_id").
		WithDescription("Collection schema root").
		WithSchema(openapi3.NewStringSchema())
	collectionVersionIdQueryParam := openapi3.NewQueryParameter("version_id").
		WithDescription("Collection schema version id").
		WithSchema(openapi3.NewStringSchema())
	collectionGetInactiveQueryParam := openapi3.NewQueryParameter("get_inactive").
		WithDescription("If true, inactive collections will be returned in addition to active ones").
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
	collectionDescribe.AddParameter(collectionGetInactiveQueryParam)
	collectionDescribe.AddResponse(200, collectionsResponse)
	collectionDescribe.Responses.Set("400", errorResponse)

	viewRefresh := openapi3.NewOperation()
	viewRefresh.OperationID = "view_refresh"
	viewRefresh.Description = "Refresh view(s) by name, schema id, or version id."
	viewRefresh.Tags = []string{"view"}
	viewRefresh.AddParameter(collectionNameQueryParam)
	viewRefresh.AddParameter(collectionSchemaRootQueryParam)
	viewRefresh.AddParameter(collectionVersionIdQueryParam)
	viewRefresh.AddParameter(collectionGetInactiveQueryParam)
	viewRefresh.Responses = openapi3.NewResponses()
	viewRefresh.Responses.Set("200", successResponse)
	viewRefresh.Responses.Set("400", errorResponse)

	patchCollectionRequest := openapi3.NewRequestBody().
		WithJSONSchemaRef(patchCollectionRequestSchema)

	patchCollection := openapi3.NewOperation()
	patchCollection.OperationID = "patch_collection"
	patchCollection.Description = "Update collection definitions"
	patchCollection.Tags = []string{"collection"}
	patchCollection.RequestBody = &openapi3.RequestBodyRef{
		Value: patchCollectionRequest,
	}
	patchCollection.Responses = openapi3.NewResponses()
	patchCollection.Responses.Set("200", successResponse)
	patchCollection.Responses.Set("400", errorResponse)

	addViewResponseSchema := openapi3.NewOneOfSchema()
	addViewResponseSchema.OneOf = openapi3.SchemaRefs{
		collectionSchema,
		openapi3.NewSchemaRef("", collectionsSchema),
	}

	addViewResponse := openapi3.NewResponse().
		WithDescription("The created collection and embedded schemas for the added view.").
		WithJSONSchema(addViewResponseSchema)

	addViewRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithJSONSchemaRef(addViewSchema)

	views := openapi3.NewOperation()
	views.OperationID = "view"
	views.Description = "Manage database views."
	views.Tags = []string{"view"}
	views.RequestBody = &openapi3.RequestBodyRef{
		Value: addViewRequest,
	}
	views.AddResponse(200, addViewResponse)
	views.Responses.Set("400", errorResponse)

	setMigrationRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithJSONSchemaRef(lensConfigSchema)

	setMigrationSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/set_migration",
	}

	setMigrationResponse := openapi3.NewResponse().
		WithDescription("Lens info").
		WithJSONSchemaRef(setMigrationSchema)
	setMigration := openapi3.NewOperation()
	setMigration.OperationID = "collection_set_migration"
	setMigration.Description = "Set a lens migration between collection versions"
	setMigration.Tags = []string{"collection"}
	setMigration.RequestBody = &openapi3.RequestBodyRef{
		Value: setMigrationRequest,
	}
	setMigration.Responses = openapi3.NewResponses()
	setMigration.AddResponse(200, setMigrationResponse)
	setMigration.Responses.Set("400", errorResponse)

	addLensRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/add_lens_request",
	}
	addLensRequestBody := openapi3.NewRequestBody().
		WithRequired(true).
		WithJSONSchemaRef(addLensRequestSchema)

	addLensResponseSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/add_lens_response",
	}
	addLensResponse := openapi3.NewResponse().
		WithDescription("Lens CID").
		WithJSONSchemaRef(addLensResponseSchema)

	addLens := openapi3.NewOperation()
	addLens.OperationID = "lens_add"
	addLens.Description = "Add a lens to the lens store"
	addLens.Tags = []string{"lens"}
	addLens.RequestBody = &openapi3.RequestBodyRef{
		Value: addLensRequestBody,
	}
	addLens.Responses = openapi3.NewResponses()
	addLens.AddResponse(200, addLensResponse)
	addLens.Responses.Set("400", errorResponse)

	listLensesResponseSchema := &openapi3.SchemaRef{
		Value: openapi3.NewObjectSchema().
			WithProperty("lenses", openapi3.NewObjectSchema()),
	}
	listLensesResponse := openapi3.NewResponse().
		WithDescription("List of stored lenses").
		WithJSONSchemaRef(listLensesResponseSchema)

	listLenses := openapi3.NewOperation()
	listLenses.OperationID = "lens_list"
	listLenses.Description = "List all stored lenses"
	listLenses.Tags = []string{"lens"}
	listLenses.Responses = openapi3.NewResponses()
	listLenses.AddResponse(200, listLensesResponse)
	listLenses.Responses.Set("400", errorResponse)

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
	graphQLPost.Responses.Set("400", errorResponse)

	graphQLQueryParam := openapi3.NewQueryParameter("query").
		WithSchema(openapi3.NewStringSchema())

	graphQLGet := openapi3.NewOperation()
	graphQLGet.Description = "GraphQL GET endpoint"
	graphQLGet.OperationID = "graphql_get"
	graphQLGet.Tags = []string{"graphql"}
	graphQLGet.AddParameter(graphQLQueryParam)
	graphQLGet.AddResponse(200, graphQLResponse)
	graphQLGet.Responses.Set("400", errorResponse)

	debugDump := openapi3.NewOperation()
	debugDump.Description = "Dump database"
	debugDump.OperationID = "debug_dump"
	debugDump.Tags = []string{"debug"}
	debugDump.Responses = openapi3.NewResponses()
	debugDump.Responses.Set("200", successResponse)
	debugDump.Responses.Set("400", errorResponse)

	identityResponse := openapi3.NewResponse().
		WithDescription("Identity").
		WithJSONSchemaRef(identitySchema)

	nodeIdentity := openapi3.NewOperation()
	nodeIdentity.OperationID = "node_identity"
	nodeIdentity.Description = "Get node's public identity"
	nodeIdentity.Tags = []string{"node", "identity"}
	nodeIdentity.AddResponse(200, identityResponse)
	nodeIdentity.Responses.Set("400", errorResponse)

	indexSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/index",
	}
	indexArraySchema := openapi3.NewArraySchema()
	indexArraySchema.Items = indexSchema

	listIndexesMapSchema := openapi3.NewObjectSchema()
	listIndexesMapSchema.AdditionalProperties = openapi3.AdditionalProperties{
		Schema: openapi3.NewSchemaRef("", indexArraySchema),
	}

	listIndexesResponse := openapi3.NewResponse().
		WithDescription("Map of collection names to their indexes").
		WithJSONSchema(listIndexesMapSchema)

	listIndexes := openapi3.NewOperation()
	listIndexes.OperationID = "indexes_list_all"
	listIndexes.Description = "List all indexes for all collections"
	listIndexes.Tags = []string{"index"}
	listIndexes.AddResponse(200, listIndexesResponse)
	listIndexes.Responses.Set("400", errorResponse)

	encryptedIndexSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/encrypted_index",
	}
	encryptedIndexArraySchema := openapi3.NewArraySchema()
	encryptedIndexArraySchema.Items = encryptedIndexSchema

	listEncryptedIndexesMapSchema := openapi3.NewObjectSchema()
	listEncryptedIndexesMapSchema.AdditionalProperties = openapi3.AdditionalProperties{
		Schema: openapi3.NewSchemaRef("", encryptedIndexArraySchema),
	}

	listEncryptedIndexesResponse := openapi3.NewResponse().
		WithDescription("Map of collection names to their encrypted indexes").
		WithJSONSchema(listEncryptedIndexesMapSchema)

	listEncryptedIndexes := openapi3.NewOperation()
	listEncryptedIndexes.OperationID = "encrypted_indexes_list_all"
	listEncryptedIndexes.Description = "List all encrypted indexes for all collections"
	listEncryptedIndexes.Tags = []string{"encrypted_index"}
	listEncryptedIndexes.AddResponse(200, listEncryptedIndexesResponse)
	listEncryptedIndexes.Responses.Set("400", errorResponse)

	router.AddRoute("/backup/export", http.MethodPost, backupExport, h.BasicExport)
	router.AddRoute("/backup/import", http.MethodPost, backupImport, h.BasicImport)
	router.AddRoute("/collections", http.MethodGet, collectionDescribe, h.GetCollection)
	router.AddRoute("/collections", http.MethodPatch, patchCollection, h.PatchCollection)
	router.AddRoute("/collections/indexes", http.MethodGet, listIndexes, h.ListIndexes)
	router.AddRoute("/encrypted-indexes", http.MethodGet, listEncryptedIndexes, h.ListAllEncryptedIndexes)
	router.AddRoute("/collections/default", http.MethodPost, setActiveCollectionVersion, h.SetActiveCollectionVersion)
	router.AddRoute("/collections/migrations", http.MethodPost, setMigration, h.SetMigration)
	router.AddRoute("/view", http.MethodPost, views, h.AddView)
	router.AddRoute("/view/refresh", http.MethodPost, viewRefresh, h.RefreshViews)
	router.AddRoute("/graphql", http.MethodGet, graphQLGet, h.ExecRequest)
	router.AddRoute("/graphql", http.MethodPost, graphQLPost, h.ExecRequest)
	router.AddRoute("/debug/dump", http.MethodGet, debugDump, h.PrintDump)
	router.AddRoute("/schema", http.MethodPost, addSchema, h.AddSchema)
	router.AddRoute("/lens", http.MethodPost, addLens, h.AddLens)
	router.AddRoute("/lens", http.MethodGet, listLenses, h.ListLenses)
	router.AddRoute("/node/identity", http.MethodGet, nodeIdentity, h.GetNodeIdentity)
}

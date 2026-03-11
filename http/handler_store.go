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
	"github.com/sourcenetwork/defradb/internal/datastore"
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

func (h *storeHandler) AddCollection(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()

	txn, hadTxn := datastore.CtxTryGetClientTxn(ctx)

	sdl, err := io.ReadAll(req.Body)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	opt := options.WithIdentity(options.AddCollection(), identity.FromContext(ctx))

	// If there is an explicit transaction, use it. Otherwise use the db.
	var cols []client.CollectionVersion
	if !hadTxn {
		cols, err = db.AddCollection(ctx, string(sdl), opt)
	} else {
		cols, err = txn.AddCollection(ctx, string(sdl), opt)
	}

	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	responseJSON(rw, http.StatusOK, cols)
}

func (h *storeHandler) PatchCollection(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()

	txn, hadTxn := datastore.CtxTryGetClientTxn(ctx)

	var message patchCollectionRequest
	err := requestJSON(req, &message)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	opt := options.WithIdentity(options.PatchCollection(), identity.FromContext(ctx))

	// If there is an explicit transaction, use it. Otherwise use the db.
	if !hadTxn {
		err = db.PatchCollection(ctx, message.Patch, message.Migration, opt)
	} else {
		err = txn.PatchCollection(ctx, message.Patch, message.Migration, opt)
	}
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (h *storeHandler) SetActiveCollectionVersion(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()

	txn, hadTxn := datastore.CtxTryGetClientTxn(ctx)

	collectionVersionID, err := io.ReadAll(req.Body)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	opt := options.WithIdentity(options.SetActiveCollectionVersion(), identity.FromContext(ctx))
	// If there is an explicit transaction, use it. Otherwise use the db.
	if !hadTxn {
		err = db.SetActiveCollectionVersion(ctx, string(collectionVersionID), opt)
	} else {
		err = txn.SetActiveCollectionVersion(ctx, string(collectionVersionID), opt)
	}

	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (h *storeHandler) AddView(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()

	txn, hadTxn := datastore.CtxTryGetClientTxn(ctx)

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

	// If there is an explicit transaction, use it. Otherwise use the db.
	var defs []client.CollectionVersion
	if !hadTxn {
		defs, err = db.AddView(ctx, message.Query, message.SDL, opt)
	} else {
		defs, err = txn.AddView(ctx, message.Query, message.SDL, opt)
	}

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

	ctx := req.Context()
	txn, hadTxn := datastore.CtxTryGetClientTxn(ctx)

	var cfg client.LensConfig
	if err := requestJSON(req, &cfg); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	opts := options.WithIdentity(options.SetMigration(), identity.FromContext(ctx))

	// If there is an explicit transaction, use it. Otherwise use the db.
	var lensID string
	var err error
	if !hadTxn {
		lensID, err = db.SetMigration(ctx, cfg, opts)
	} else {
		lensID, err = txn.SetMigration(ctx, cfg, opts)
	}

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

	ctx := req.Context()
	txn, hadTxn := datastore.CtxTryGetClientTxn(ctx)

	var addLensReq AddLensRequest
	if err := requestJSON(req, &addLensReq); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	opts := options.WithIdentity(options.AddLens(), identity.FromContext(ctx))

	// If there is an explicit transaction, use it. Otherwise use the db.
	var lensID string
	var err error
	if !hadTxn {
		lensID, err = db.AddLens(ctx, addLensReq.Lens, opts)
	} else {
		lensID, err = txn.AddLens(ctx, addLensReq.Lens, opts)
	}
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

	ctx := req.Context()
	txn, hadTxn := datastore.CtxTryGetClientTxn(ctx)

	opts := options.WithIdentity(options.ListLenses(), identity.FromContext(ctx))

	// If there is an explicit transaction, use it. Otherwise use the db.
	var lenses map[string]model.Lens
	var err error
	if !hadTxn {
		lenses, err = db.ListLenses(ctx, opts)
	} else {
		lenses, err = txn.ListLenses(ctx, opts)
	}
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	responseJSON(rw, http.StatusOK, &ListLensesResponse{Lenses: lenses})
}

func (h *storeHandler) GetCollection(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()

	txn, hadTxn := datastore.CtxTryGetClientTxn(ctx)

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

	var cols []client.Collection
	var err error
	if !hadTxn {
		cols, err = db.GetCollections(ctx, opt)
	} else {
		cols, err = txn.GetCollections(ctx, opt)
	}
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

	ctx := req.Context()
	txn, hadTxn := datastore.CtxTryGetClientTxn(ctx)

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

	// If there is an explicit transaction, use it. Otherwise use the db.
	var err error
	if !hadTxn {
		err = db.RefreshViews(ctx, opt)
	} else {
		err = txn.RefreshViews(ctx, opt)
	}

	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (h *storeHandler) ListIndexes(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	txn, hadTxn := datastore.CtxTryGetClientTxn(req.Context())

	// If there is an explicit transaction, use it. Otherwise use the db.
	var indexes map[client.CollectionName][]client.IndexDescription
	var err error
	if !hadTxn {
		indexes, err = db.ListIndexes(req.Context())
	} else {
		indexes, err = txn.ListIndexes(req.Context())
	}
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	responseJSON(rw, http.StatusOK, indexes)
}

func (h *storeHandler) ListAllEncryptedIndexes(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()
	txn, hadTxn := datastore.CtxTryGetClientTxn(ctx)

	opts := options.WithIdentity(options.ListAllEncryptedIndexes(), identity.FromContext(ctx))

	// If there is an explicit transaction, use it. Otherwise use the db.
	var indexes map[client.CollectionName][]client.EncryptedIndexDescription
	var err error
	if !hadTxn {
		indexes, err = db.ListAllEncryptedIndexes(ctx, opts)
	} else {
		indexes, err = txn.ListAllEncryptedIndexes(ctx, opts)
	}

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

	txn, hadTxn := datastore.CtxTryGetClientTxn(ctx)

	request, opts, err := extractGraphQLRequest(req)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	opts = options.WithIdentity(opts, identity.FromContext(ctx))
	var result *client.RequestResult
	if !hadTxn {
		result = db.ExecRequest(ctx, request.Query, opts)
	} else {
		result = txn.ExecRequest(ctx, request.Query, opts)
	}

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
		Ref: "#/components/schemas/request_graphql",
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

	addCollectionResponse := openapi3.NewResponse().
		WithDescription("Collection(s)").
		WithJSONSchema(collectionArraySchema)

	addCollectionRequest := openapi3.NewRequestBody().
		WithContent(openapi3.NewContentWithSchema(openapi3.NewStringSchema(), []string{"text/plain"}))

	addCollection := openapi3.NewOperation()
	addCollection.OperationID = "add_collection"
	addCollection.Description = "Add a new collection"
	addCollection.Tags = []string{"collection"}
	addCollection.RequestBody = &openapi3.RequestBodyRef{
		Value: addCollectionRequest,
	}
	addCollection.AddResponse(200, addCollectionResponse)
	addCollection.Responses.Set("400", errorResponse)

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

	exportBackup := openapi3.NewOperation()
	exportBackup.OperationID = "export_backup"
	exportBackup.Description = "Export a database backup to file"
	exportBackup.Tags = []string{"backup"}
	exportBackup.Responses = openapi3.NewResponses()
	exportBackup.Responses.Set("200", successResponse)
	exportBackup.Responses.Set("400", errorResponse)
	exportBackup.RequestBody = &openapi3.RequestBodyRef{
		Value: backupRequest,
	}

	importBackup := openapi3.NewOperation()
	importBackup.OperationID = "import_backup"
	importBackup.Description = "Import a database backup from file"
	importBackup.Tags = []string{"backup"}
	importBackup.Responses = openapi3.NewResponses()
	importBackup.Responses.Set("200", successResponse)
	importBackup.Responses.Set("400", errorResponse)
	importBackup.RequestBody = &openapi3.RequestBodyRef{
		Value: backupRequest,
	}

	collectionNameQueryParam := openapi3.NewQueryParameter("name").
		WithDescription("Collection name").
		WithSchema(openapi3.NewStringSchema())
	collectionIDQueryParam := openapi3.NewQueryParameter("collection_id").
		WithDescription("Collection ID").
		WithSchema(openapi3.NewStringSchema())
	collectionVersionIdQueryParam := openapi3.NewQueryParameter("version_id").
		WithDescription("Collection version ID").
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
		WithDescription("Collection(s) with matching name, collection id, or version id.").
		WithJSONSchema(collectionResponseSchema)

	describeCollection := openapi3.NewOperation()
	describeCollection.OperationID = "describe_collection"
	describeCollection.Description = "Introspect collection(s) by name, collection id, or version id."
	describeCollection.Tags = []string{"collection"}
	describeCollection.AddParameter(collectionNameQueryParam)
	describeCollection.AddParameter(collectionIDQueryParam)
	describeCollection.AddParameter(collectionVersionIdQueryParam)
	describeCollection.AddParameter(collectionGetInactiveQueryParam)
	describeCollection.AddResponse(200, collectionsResponse)
	describeCollection.Responses.Set("400", errorResponse)

	refreshView := openapi3.NewOperation()
	refreshView.OperationID = "refresh_view"
	refreshView.Description = "Refresh view(s) by name, collection id, or version id."
	refreshView.Tags = []string{"view"}
	refreshView.AddParameter(collectionNameQueryParam)
	refreshView.AddParameter(collectionIDQueryParam)
	refreshView.AddParameter(collectionVersionIdQueryParam)
	refreshView.AddParameter(collectionGetInactiveQueryParam)
	refreshView.Responses = openapi3.NewResponses()
	refreshView.Responses.Set("200", successResponse)
	refreshView.Responses.Set("400", errorResponse)

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
		WithDescription("The added collection and embedded schemas for the added view.").
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
	setMigration.OperationID = "set_collection_migration"
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
	addLens.OperationID = "add_lens"
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
	listLenses.OperationID = "list_lenses"
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

	postGraphQL := openapi3.NewOperation()
	postGraphQL.Description = "GraphQL POST endpoint"
	postGraphQL.OperationID = "post_graphql"
	postGraphQL.Tags = []string{"graphql"}
	postGraphQL.RequestBody = &openapi3.RequestBodyRef{
		Value: graphQLRequest,
	}
	postGraphQL.AddResponse(200, graphQLResponse)
	postGraphQL.Responses.Set("400", errorResponse)

	graphQLQueryParam := openapi3.NewQueryParameter("query").
		WithSchema(openapi3.NewStringSchema())

	getGraphQL := openapi3.NewOperation()
	getGraphQL.Description = "GraphQL GET endpoint"
	getGraphQL.OperationID = "get_graphql"
	getGraphQL.Tags = []string{"graphql"}
	getGraphQL.AddParameter(graphQLQueryParam)
	getGraphQL.AddResponse(200, graphQLResponse)
	getGraphQL.Responses.Set("400", errorResponse)

	debugDump := openapi3.NewOperation()
	debugDump.Description = "Dump database"
	debugDump.OperationID = "dump_debug"
	debugDump.Tags = []string{"debug"}
	debugDump.Responses = openapi3.NewResponses()
	debugDump.Responses.Set("200", successResponse)
	debugDump.Responses.Set("400", errorResponse)

	identityResponse := openapi3.NewResponse().
		WithDescription("Identity").
		WithJSONSchemaRef(identitySchema)

	nodeIdentity := openapi3.NewOperation()
	nodeIdentity.OperationID = "get_node_identity"
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
	listIndexes.OperationID = "list_all_indexes"
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
	listEncryptedIndexes.OperationID = "list_all_encrypted_indexes"
	listEncryptedIndexes.Description = "List all encrypted indexes for all collections"
	listEncryptedIndexes.Tags = []string{"encrypted_index"}
	listEncryptedIndexes.AddResponse(200, listEncryptedIndexesResponse)
	listEncryptedIndexes.Responses.Set("400", errorResponse)

	router.AddRoute("/backup/export", http.MethodPost, exportBackup, h.BasicExport)
	router.AddRoute("/backup/import", http.MethodPost, importBackup, h.BasicImport)
	router.AddRoute("/collections", http.MethodGet, describeCollection, h.GetCollection)
	router.AddRoute("/collections", http.MethodPatch, patchCollection, h.PatchCollection)
	router.AddRoute("/collections/indexes", http.MethodGet, listIndexes, h.ListIndexes)
	router.AddRoute("/encrypted-indexes", http.MethodGet, listEncryptedIndexes, h.ListAllEncryptedIndexes)
	router.AddRoute("/collections/default", http.MethodPost, setActiveCollectionVersion, h.SetActiveCollectionVersion)
	router.AddRoute("/collections/migrations", http.MethodPost, setMigration, h.SetMigration)
	router.AddRoute("/view", http.MethodPost, views, h.AddView)
	router.AddRoute("/view/refresh", http.MethodPost, refreshView, h.RefreshViews)
	router.AddRoute("/graphql", http.MethodGet, getGraphQL, h.ExecRequest)
	router.AddRoute("/graphql", http.MethodPost, postGraphQL, h.ExecRequest)
	router.AddRoute("/debug/dump", http.MethodGet, debugDump, h.PrintDump)
	router.AddRoute("/collections", http.MethodPost, addCollection, h.AddCollection)
	router.AddRoute("/lens", http.MethodPost, addLens, h.AddLens)
	router.AddRoute("/lens", http.MethodGet, listLenses, h.ListLenses)
	router.AddRoute("/node/identity", http.MethodGet, nodeIdentity, h.GetNodeIdentity)
}

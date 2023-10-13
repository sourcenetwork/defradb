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

func (s *storeHandler) SetReplicator(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	var rep client.Replicator
	if err := requestJSON(req, &rep); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	err := store.SetReplicator(req.Context(), rep)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *storeHandler) DeleteReplicator(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	var rep client.Replicator
	if err := requestJSON(req, &rep); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	err := store.DeleteReplicator(req.Context(), rep)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *storeHandler) GetAllReplicators(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	reps, err := store.GetAllReplicators(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, reps)
}

func (s *storeHandler) AddP2PCollection(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	var collectionIDs []string
	if err := requestJSON(req, &collectionIDs); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	err := store.AddP2PCollections(req.Context(), collectionIDs)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *storeHandler) RemoveP2PCollection(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	var collectionIDs []string
	if err := requestJSON(req, &collectionIDs); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	err := store.RemoveP2PCollections(req.Context(), collectionIDs)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *storeHandler) GetAllP2PCollections(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	cols, err := store.GetAllP2PCollections(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, cols)
}

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
		responseJSON(rw, http.StatusOK, col.Description())
	case req.URL.Query().Has("schema_id"):
		col, err := store.GetCollectionBySchemaID(req.Context(), req.URL.Query().Get("schema_id"))
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		responseJSON(rw, http.StatusOK, col.Description())
	case req.URL.Query().Has("version_id"):
		col, err := store.GetCollectionByVersionID(req.Context(), req.URL.Query().Get("version_id"))
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		responseJSON(rw, http.StatusOK, col.Description())
	default:
		cols, err := store.GetAllCollections(req.Context())
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		colDesc := make([]client.CollectionDescription, len(cols))
		for i, col := range cols {
			colDesc[i] = col.Description()
		}
		responseJSON(rw, http.StatusOK, colDesc)
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

type PeerInfoResponse struct {
	PeerID string `json:"peerID"`
}

func (s *storeHandler) PeerInfo(rw http.ResponseWriter, req *http.Request) {
	var res PeerInfoResponse
	if value, ok := req.Context().Value(peerIdContextKey).(string); ok {
		res.PeerID = value
	}
	responseJSON(rw, http.StatusOK, &res)
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
	graphQLRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/graphql_request",
	}
	graphQLResponseSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/graphql_response",
	}
	backupConfigSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/backup_config",
	}
	peerInfoSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/peer_info",
	}
	replicatorSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/replicator",
	}

	backupRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithJSONSchemaRef(backupConfigSchema)

	backupExport := openapi3.NewOperation()
	backupExport.OperationID = "backup_export"
	backupExport.Responses = make(openapi3.Responses)
	backupExport.Responses["200"] = successResponse
	backupExport.Responses["400"] = errorResponse
	backupExport.RequestBody = &openapi3.RequestBodyRef{
		Value: backupRequest,
	}

	backupImport := openapi3.NewOperation()
	backupImport.OperationID = "backup_import"
	backupImport.Responses = make(openapi3.Responses)
	backupImport.Responses["200"] = successResponse
	backupImport.Responses["400"] = errorResponse
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
	collectionDescribe.AddParameter(collectionNameQueryParam)
	collectionDescribe.AddParameter(collectionSchemaIdQueryParam)
	collectionDescribe.AddParameter(collectionVersionIdQueryParam)
	collectionDescribe.AddResponse(200, collectionsResponse)
	collectionDescribe.Responses["400"] = errorResponse

	graphQLRequest := openapi3.NewRequestBody().
		WithContent(openapi3.NewContentWithJSONSchemaRef(graphQLRequestSchema))

	graphQLResponse := openapi3.NewResponse().
		WithDescription("GraphQL response").
		WithContent(openapi3.NewContentWithJSONSchemaRef(graphQLResponseSchema))

	graphQLPost := openapi3.NewOperation()
	graphQLPost.Description = "GraphQL POST endpoint"
	graphQLPost.OperationID = "graphql_post"
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
	graphQLGet.AddParameter(graphQLQueryParam)
	graphQLGet.AddResponse(200, graphQLResponse)
	graphQLGet.Responses["400"] = errorResponse

	debugDump := openapi3.NewOperation()
	debugDump.Description = "Dump database"
	debugDump.OperationID = "debug_dump"
	debugDump.Responses = make(openapi3.Responses)
	debugDump.Responses["200"] = successResponse
	debugDump.Responses["400"] = errorResponse

	peerInfoResponse := openapi3.NewResponse().
		WithDescription("Peer network info").
		WithContent(openapi3.NewContentWithJSONSchemaRef(peerInfoSchema))

	peerInfo := openapi3.NewOperation()
	peerInfo.OperationID = "peer_info"
	peerInfo.AddResponse(200, peerInfoResponse)
	peerInfo.Responses["400"] = errorResponse

	getReplicatorsSchema := openapi3.NewArraySchema()
	getReplicatorsSchema.Items = replicatorSchema
	getReplicatorsResponse := openapi3.NewResponse().
		WithDescription("Replicators").
		WithContent(openapi3.NewContentWithJSONSchema(getReplicatorsSchema))

	getReplicators := openapi3.NewOperation()
	getReplicators.Description = "List peer replicators"
	getReplicators.OperationID = "peer_replicator_list"
	getReplicators.AddResponse(200, getReplicatorsResponse)
	getReplicators.Responses["400"] = errorResponse

	replicatorRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(replicatorSchema))

	setReplicator := openapi3.NewOperation()
	setReplicator.Description = "Add peer replicators"
	setReplicator.OperationID = "peer_replicator_set"
	setReplicator.RequestBody = &openapi3.RequestBodyRef{
		Value: replicatorRequest,
	}
	setReplicator.Responses = make(openapi3.Responses)
	setReplicator.Responses["200"] = successResponse
	setReplicator.Responses["400"] = errorResponse

	deleteReplicator := openapi3.NewOperation()
	deleteReplicator.Description = "Delete peer replicators"
	deleteReplicator.OperationID = "peer_replicator_delete"
	deleteReplicator.RequestBody = &openapi3.RequestBodyRef{
		Value: replicatorRequest,
	}
	deleteReplicator.Responses = make(openapi3.Responses)
	deleteReplicator.Responses["200"] = successResponse
	deleteReplicator.Responses["400"] = errorResponse

	peerCollectionsSchema := openapi3.NewArraySchema().
		WithItems(openapi3.NewStringSchema())

	peerCollectionRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchema(peerCollectionsSchema))

	getPeerCollectionsResponse := openapi3.NewResponse().
		WithDescription("Peer collections").
		WithContent(openapi3.NewContentWithJSONSchema(peerCollectionsSchema))

	getPeerCollections := openapi3.NewOperation()
	getPeerCollections.Description = "List peer collections"
	getPeerCollections.OperationID = "peer_collection_list"
	getPeerCollections.AddResponse(200, getPeerCollectionsResponse)
	getPeerCollections.Responses["400"] = errorResponse

	addPeerCollections := openapi3.NewOperation()
	addPeerCollections.Description = "Add peer collections"
	addPeerCollections.OperationID = "peer_collection_add"
	addPeerCollections.RequestBody = &openapi3.RequestBodyRef{
		Value: peerCollectionRequest,
	}
	addPeerCollections.Responses = make(openapi3.Responses)
	addPeerCollections.Responses["200"] = successResponse
	addPeerCollections.Responses["400"] = errorResponse

	removePeerCollections := openapi3.NewOperation()
	removePeerCollections.Description = "Remove peer collections"
	removePeerCollections.OperationID = "peer_collection_remove"
	removePeerCollections.RequestBody = &openapi3.RequestBodyRef{
		Value: peerCollectionRequest,
	}
	removePeerCollections.Responses = make(openapi3.Responses)
	removePeerCollections.Responses["200"] = successResponse
	removePeerCollections.Responses["400"] = errorResponse

	router.AddRoute("/p2p/info", http.MethodGet, peerInfo, h.PeerInfo)
	router.AddRoute("/p2p/replicators", http.MethodGet, getReplicators, h.GetAllReplicators)
	router.AddRoute("/p2p/replicators", http.MethodPost, setReplicator, h.SetReplicator)
	router.AddRoute("/p2p/replicators", http.MethodDelete, deleteReplicator, h.DeleteReplicator)
	router.AddRoute("/p2p/collections", http.MethodGet, getPeerCollections, h.GetAllP2PCollections)
	router.AddRoute("/p2p/collections", http.MethodPost, addPeerCollections, h.AddP2PCollection)
	router.AddRoute("/p2p/collections", http.MethodDelete, removePeerCollections, h.RemoveP2PCollection)
	router.AddRoute("/backup/export", http.MethodPost, backupExport, h.BasicExport)
	router.AddRoute("/backup/import", http.MethodPost, backupImport, h.BasicImport)
	router.AddRoute("/collections", http.MethodGet, collectionDescribe, h.GetCollection)
	router.AddRoute("/graphql", http.MethodGet, graphQLGet, h.ExecRequest)
	router.AddRoute("/graphql", http.MethodPost, graphQLPost, h.ExecRequest)
	router.AddRoute("/debug/dump", http.MethodGet, debugDump, h.PrintDump)
}

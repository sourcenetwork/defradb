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
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/encryption"
)

const docEncryptParam = "encrypt"
const docEncryptFieldsParam = "encryptFields"

type collectionHandler struct{}

type CollectionDeleteRequest struct {
	Filter any `json:"filter"`
}

type CollectionUpdateRequest struct {
	Filter  any    `json:"filter"`
	Updater string `json:"updater"`
}

func (s *collectionHandler) Create(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	data, err := io.ReadAll(req.Body)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	ctx := req.Context()
	q := req.URL.Query()
	encConf := encryption.DocEncConfig{}
	if q.Get(docEncryptParam) == "true" {
		encConf.IsDocEncrypted = true
	}
	if q.Get(docEncryptFieldsParam) != "" {
		encConf.EncryptedFields = strings.Split(q.Get(docEncryptFieldsParam), ",")
	}
	if encConf.IsDocEncrypted || len(encConf.EncryptedFields) > 0 {
		ctx = encryption.SetContextConfig(ctx, encConf)
	}

	switch {
	case client.IsJSONArray(data):
		docList, err := client.NewDocsFromJSON(data, col.Definition())
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}

		if err := col.CreateMany(ctx, docList); err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		rw.WriteHeader(http.StatusOK)
	default:
		doc, err := client.NewDocFromJSON(data, col.Definition())
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		if err := col.Create(ctx, doc); err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		rw.WriteHeader(http.StatusOK)
	}
}

func (s *collectionHandler) DeleteWithFilter(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	var request CollectionDeleteRequest
	if err := requestJSON(req, &request); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	result, err := col.DeleteWithFilter(req.Context(), request.Filter)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, result)
}

func (s *collectionHandler) UpdateWithFilter(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	var request CollectionUpdateRequest
	if err := requestJSON(req, &request); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	result, err := col.UpdateWithFilter(req.Context(), request.Filter, request.Updater)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, result)
}

func (s *collectionHandler) Update(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	docID, err := client.NewDocIDFromString(chi.URLParam(req, "docID"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	doc, err := col.Get(req.Context(), docID, true)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	if doc == nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{client.ErrDocumentNotFoundOrNotAuthorized})
		return
	}

	patch, err := io.ReadAll(req.Body)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	if err := doc.SetWithJSON(patch); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	err = col.Update(req.Context(), doc)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *collectionHandler) Delete(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	docID, err := client.NewDocIDFromString(chi.URLParam(req, "docID"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	_, err = col.Delete(req.Context(), docID)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *collectionHandler) Get(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)
	showDeleted, _ := strconv.ParseBool(req.URL.Query().Get("show_deleted"))

	docID, err := client.NewDocIDFromString(chi.URLParam(req, "docID"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	doc, err := col.Get(req.Context(), docID, showDeleted)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	if doc == nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{client.ErrDocumentNotFoundOrNotAuthorized})
		return
	}

	docMap, err := doc.ToMap()
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, docMap)
}

type DocIDResult struct {
	DocID string `json:"docID"`
	Error string `json:"error"`
}

func (s *collectionHandler) GetAllDocIDs(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	flusher, ok := rw.(http.Flusher)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrStreamingNotSupported})
		return
	}

	docIDsResult, err := col.GetAllDocIDs(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")

	rw.WriteHeader(http.StatusOK)
	flusher.Flush()

	for docID := range docIDsResult {
		results := &DocIDResult{
			DocID: docID.ID.String(),
		}
		if docID.Err != nil {
			results.Error = docID.Err.Error()
		}
		data, err := json.Marshal(results)
		if err != nil {
			return
		}
		_, err = fmt.Fprintf(rw, "data: %s\n\n", data)
		if err != nil {
			return
		}
		flusher.Flush()
	}
}

func (s *collectionHandler) CreateIndex(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	var indexDesc client.IndexDescription
	if err := requestJSON(req, &indexDesc); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	index, err := col.CreateIndex(req.Context(), indexDesc)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, index)
}

func (s *collectionHandler) GetIndexes(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(dbContextKey).(client.Store)
	indexesMap, err := store.GetAllIndexes(req.Context())

	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	indexes := make([]client.IndexDescription, 0, len(indexesMap))
	for _, index := range indexesMap {
		indexes = append(indexes, index...)
	}
	responseJSON(rw, http.StatusOK, indexes)
}

func (s *collectionHandler) DropIndex(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	err := col.DropIndex(req.Context(), chi.URLParam(req, "index"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h *collectionHandler) bindRoutes(router *Router) {
	errorResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/error",
	}
	successResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/success",
	}
	collectionUpdateSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/collection_update",
	}
	updateResultSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/update_result",
	}
	collectionDeleteSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/collection_delete",
	}
	deleteResultSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/delete_result",
	}
	documentSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/document",
	}
	indexSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/index",
	}

	collectionNamePathParam := openapi3.NewPathParameter("name").
		WithDescription("Collection name").
		WithRequired(true).
		WithSchema(openapi3.NewStringSchema())

	documentArraySchema := openapi3.NewArraySchema()
	documentArraySchema.Items = documentSchema

	collectionCreateSchema := openapi3.NewOneOfSchema()
	collectionCreateSchema.OneOf = openapi3.SchemaRefs{
		documentSchema,
		openapi3.NewSchemaRef("", documentArraySchema),
	}

	collectionCreateRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchema(collectionCreateSchema))

	collectionCreate := openapi3.NewOperation()
	collectionCreate.OperationID = "collection_create"
	collectionCreate.Description = "Create document(s) in a collection"
	collectionCreate.Tags = []string{"collection"}
	collectionCreate.AddParameter(collectionNamePathParam)
	collectionCreate.RequestBody = &openapi3.RequestBodyRef{
		Value: collectionCreateRequest,
	}
	collectionCreate.Responses = openapi3.NewResponses()
	collectionCreate.Responses.Set("200", successResponse)
	collectionCreate.Responses.Set("400", errorResponse)

	collectionUpdateWithRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(collectionUpdateSchema))

	collectionUpdateWithResponse := openapi3.NewResponse().
		WithDescription("Update results").
		WithJSONSchemaRef(updateResultSchema)

	collectionUpdateWith := openapi3.NewOperation()
	collectionUpdateWith.OperationID = "collection_update_with_filter"
	collectionUpdateWith.Description = "Update document(s) in a collection"
	collectionUpdateWith.Tags = []string{"collection"}
	collectionUpdateWith.AddParameter(collectionNamePathParam)
	collectionUpdateWith.RequestBody = &openapi3.RequestBodyRef{
		Value: collectionUpdateWithRequest,
	}
	collectionUpdateWith.AddResponse(200, collectionUpdateWithResponse)
	collectionUpdateWith.Responses.Set("400", errorResponse)

	collectionDeleteWithRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(collectionDeleteSchema))

	collectionDeleteWithResponse := openapi3.NewResponse().
		WithDescription("Delete results").
		WithJSONSchemaRef(deleteResultSchema)

	collectionDeleteWith := openapi3.NewOperation()
	collectionDeleteWith.OperationID = "collection_delete_with_filter"
	collectionDeleteWith.Description = "Delete document(s) from a collection"
	collectionDeleteWith.Tags = []string{"collection"}
	collectionDeleteWith.AddParameter(collectionNamePathParam)
	collectionDeleteWith.RequestBody = &openapi3.RequestBodyRef{
		Value: collectionDeleteWithRequest,
	}
	collectionDeleteWith.AddResponse(200, collectionDeleteWithResponse)
	collectionDeleteWith.Responses.Set("400", errorResponse)

	createIndexRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(indexSchema))
	createIndexResponse := openapi3.NewResponse().
		WithDescription("Index description").
		WithJSONSchemaRef(indexSchema)

	createIndex := openapi3.NewOperation()
	createIndex.OperationID = "index_create"
	createIndex.Description = "Create a secondary index"
	createIndex.Tags = []string{"index"}
	createIndex.AddParameter(collectionNamePathParam)
	createIndex.RequestBody = &openapi3.RequestBodyRef{
		Value: createIndexRequest,
	}
	createIndex.AddResponse(200, createIndexResponse)
	createIndex.Responses.Set("400", errorResponse)

	indexArraySchema := openapi3.NewArraySchema()
	indexArraySchema.Items = indexSchema

	getIndexesResponse := openapi3.NewResponse().
		WithDescription("List of indexes").
		WithJSONSchema(indexArraySchema)

	getIndexes := openapi3.NewOperation()
	getIndexes.OperationID = "index_list"
	getIndexes.Description = "List secondary indexes"
	getIndexes.Tags = []string{"index"}
	getIndexes.AddParameter(collectionNamePathParam)
	getIndexes.AddResponse(200, getIndexesResponse)
	getIndexes.Responses.Set("400", errorResponse)

	indexPathParam := openapi3.NewPathParameter("index").
		WithRequired(true).
		WithSchema(openapi3.NewStringSchema())

	dropIndex := openapi3.NewOperation()
	dropIndex.OperationID = "index_drop"
	dropIndex.Description = "Delete a secondary index"
	dropIndex.Tags = []string{"index"}
	dropIndex.AddParameter(collectionNamePathParam)
	dropIndex.AddParameter(indexPathParam)
	dropIndex.Responses = openapi3.NewResponses()
	dropIndex.Responses.Set("200", successResponse)
	dropIndex.Responses.Set("400", errorResponse)

	documentIDPathParam := openapi3.NewPathParameter("docID").
		WithRequired(true).
		WithSchema(openapi3.NewStringSchema())

	collectionGetResponse := openapi3.NewResponse().
		WithDescription("Document value").
		WithJSONSchemaRef(documentSchema)

	collectionGet := openapi3.NewOperation()
	collectionGet.Description = "Get a document by docID"
	collectionGet.OperationID = "collection_get"
	collectionGet.Tags = []string{"collection"}
	collectionGet.AddParameter(collectionNamePathParam)
	collectionGet.AddParameter(documentIDPathParam)
	collectionGet.AddResponse(200, collectionGetResponse)
	collectionGet.Responses.Set("400", errorResponse)

	collectionUpdate := openapi3.NewOperation()
	collectionUpdate.Description = "Update a document by docID"
	collectionUpdate.OperationID = "collection_update"
	collectionUpdate.Tags = []string{"collection"}
	collectionUpdate.AddParameter(collectionNamePathParam)
	collectionUpdate.AddParameter(documentIDPathParam)
	collectionUpdate.Responses = openapi3.NewResponses()
	collectionUpdate.Responses.Set("200", successResponse)
	collectionUpdate.Responses.Set("400", errorResponse)

	collectionDelete := openapi3.NewOperation()
	collectionDelete.Description = "Delete a document by docID"
	collectionDelete.OperationID = "collection_delete"
	collectionDelete.Tags = []string{"collection"}
	collectionDelete.AddParameter(collectionNamePathParam)
	collectionDelete.AddParameter(documentIDPathParam)
	collectionDelete.Responses = openapi3.NewResponses()
	collectionDelete.Responses.Set("200", successResponse)
	collectionDelete.Responses.Set("400", errorResponse)

	collectionKeys := openapi3.NewOperation()
	collectionKeys.AddParameter(collectionNamePathParam)
	collectionKeys.Description = "Get all document IDs"
	collectionKeys.OperationID = "collection_keys"
	collectionKeys.Tags = []string{"collection"}
	collectionKeys.Responses = openapi3.NewResponses()
	collectionKeys.Responses.Set("200", successResponse)
	collectionKeys.Responses.Set("400", errorResponse)

	router.AddRoute("/collections/{name}", http.MethodGet, collectionKeys, h.GetAllDocIDs)
	router.AddRoute("/collections/{name}", http.MethodPost, collectionCreate, h.Create)
	router.AddRoute("/collections/{name}", http.MethodPatch, collectionUpdateWith, h.UpdateWithFilter)
	router.AddRoute("/collections/{name}", http.MethodDelete, collectionDeleteWith, h.DeleteWithFilter)
	router.AddRoute("/collections/{name}/indexes", http.MethodPost, createIndex, h.CreateIndex)
	router.AddRoute("/collections/{name}/indexes", http.MethodGet, getIndexes, h.GetIndexes)
	router.AddRoute("/collections/{name}/indexes/{index}", http.MethodDelete, dropIndex, h.DropIndex)
	router.AddRoute("/collections/{name}/{docID}", http.MethodGet, collectionGet, h.Get)
	router.AddRoute("/collections/{name}/{docID}", http.MethodPatch, collectionUpdate, h.Update)
	router.AddRoute("/collections/{name}/{docID}", http.MethodDelete, collectionDelete, h.Delete)
}
